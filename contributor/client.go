package contributor

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/grokify/gogithub/auth"
)

// coAuthorRegex matches Co-authored-by trailers in commit messages.
var coAuthorRegex = regexp.MustCompile(`(?i)co-authored-by:\s*(.+?)\s*<([^>]+)>`)

// Client provides contributor profile generation functionality.
type Client struct {
	gh *github.Client
}

// NewClient creates a new contributor client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	client := auth.NewGitHubClient(context.Background(), token)
	return &Client{gh: client}, nil
}

// GenerateProfile generates a contributor profile for the given options.
func (c *Client) GenerateProfile(ctx context.Context, opts ProfileOptions) (*Profile, error) {
	// Get user info
	user, _, err := c.gh.Users.Get(ctx, opts.Username)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	profile := &Profile{
		Username:    opts.Username,
		Name:        user.GetName(),
		AvatarURL:   user.GetAvatarURL(),
		Bio:         user.GetBio(),
		Company:     user.GetCompany(),
		Location:    user.GetLocation(),
		Blog:        user.GetBlog(),
		Languages:   make(map[string]int),
		AIStats: AICollabStats{
			ByTool: make(map[string]AIToolStat),
		},
		GeneratedAt: time.Now().UTC(),
	}

	// Track AI commits across all repos
	aiCommitsByDate := make(map[string]int)

	// Set date range
	if !opts.Since.IsZero() {
		profile.DateRange.Start = opts.Since.Format("2006-01-02")
	} else {
		// Default to last year
		profile.DateRange.Start = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	}

	if !opts.Until.IsZero() {
		profile.DateRange.End = opts.Until.Format("2006-01-02")
	} else {
		profile.DateRange.End = time.Now().Format("2006-01-02")
	}

	// Get repositories the user has contributed to
	repos, err := c.getUserRepos(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("fetching repos: %w", err)
	}

	// Get contribution data for each repo
	for _, repo := range repos {
		contrib, aiData, err := c.getRepoContributions(ctx, opts.Username, repo, opts)
		if err != nil {
			// Log but continue
			continue
		}

		if contrib.Commits > 0 || contrib.PRs > 0 || contrib.Issues > 0 {
			profile.Repositories = append(profile.Repositories, contrib)
			profile.Stats.TotalCommits += contrib.Commits
			profile.Stats.TotalPRs += contrib.PRs
			profile.Stats.TotalPRsMerged += contrib.PRsMerged
			profile.Stats.TotalIssues += contrib.Issues
			profile.Stats.TotalReviews += contrib.Reviews

			if contrib.IsOwner {
				profile.Stats.OwnedRepos++
			} else {
				profile.Stats.ContributedRepos++
			}

			if contrib.Language != "" {
				profile.Languages[contrib.Language]++
			}

			// Aggregate AI stats
			for toolName, toolData := range aiData {
				existing := profile.AIStats.ByTool[toolName]
				existing.Name = toolName
				existing.Commits += toolData.Commits
				existing.Recognized = toolData.Recognized

				// Track first/last used dates
				if existing.FirstUsed == "" || toolData.FirstUsed < existing.FirstUsed {
					existing.FirstUsed = toolData.FirstUsed
				}
				if toolData.LastUsed > existing.LastUsed {
					existing.LastUsed = toolData.LastUsed
				}

				profile.AIStats.ByTool[toolName] = existing
				profile.AIStats.TotalAICommits += toolData.Commits

				// Track AI commits by date for heatmap
				for date, count := range toolData.byDate {
					aiCommitsByDate[date] += count
				}
			}
		}
	}

	profile.Stats.TotalRepositories = len(profile.Repositories)

	// Sort repositories by total contributions
	sort.Slice(profile.Repositories, func(i, j int) bool {
		totalI := profile.Repositories[i].Commits + profile.Repositories[i].PRs
		totalJ := profile.Repositories[j].Commits + profile.Repositories[j].PRs
		return totalI > totalJ
	})

	// Calculate AI stats summary
	if profile.Stats.TotalCommits > 0 {
		profile.AIStats.AICommitPercent = float64(profile.AIStats.TotalAICommits) / float64(profile.Stats.TotalCommits) * 100
	}

	// Find most used AI tool
	maxCommits := 0
	for toolName, stat := range profile.AIStats.ByTool {
		if stat.Commits > maxCommits {
			maxCommits = stat.Commits
			profile.AIStats.MostUsedTool = toolName
		}
		// Track first AI commit date
		if profile.AIStats.FirstAICommit == "" || stat.FirstUsed < profile.AIStats.FirstAICommit {
			profile.AIStats.FirstAICommit = stat.FirstUsed
		}
	}

	// Build AI activity heatmap
	var aiActivity []DailyActivity
	for date, count := range aiCommitsByDate {
		aiActivity = append(aiActivity, DailyActivity{Date: date, Count: count})
	}
	sort.Slice(aiActivity, func(i, j int) bool {
		return aiActivity[i].Date < aiActivity[j].Date
	})
	profile.AIStats.AIActivity = aiActivity

	// Get activity data for heatmap
	activity, err := c.getActivityHeatmap(ctx, opts)
	if err != nil {
		// Non-fatal, continue without activity data
		profile.Activity = []DailyActivity{}
	} else {
		profile.Activity = activity
	}

	return profile, nil
}

func (c *Client) getUserRepos(ctx context.Context, opts ProfileOptions) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	// Get user's own repos
	listOpts := &github.RepositoryListByUserOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.gh.Repositories.ListByUser(ctx, opts.Username, listOpts)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	// Filter by orgs if specified
	if len(opts.Orgs) > 0 {
		orgSet := make(map[string]bool)
		for _, org := range opts.Orgs {
			orgSet[org] = true
		}

		filtered := make([]*github.Repository, 0)
		for _, repo := range allRepos {
			if orgSet[repo.GetOwner().GetLogin()] {
				filtered = append(filtered, repo)
			}
		}
		allRepos = filtered
	}

	return allRepos, nil
}

// aiToolData tracks AI tool usage within a repo (internal type with byDate for aggregation).
type aiToolData struct {
	AIToolStat
	byDate map[string]int
}

func (c *Client) getRepoContributions(ctx context.Context, username string, repo *github.Repository, opts ProfileOptions) (RepoContrib, map[string]aiToolData, error) {
	owner := repo.GetOwner().GetLogin()
	name := repo.GetName()

	contrib := RepoContrib{
		Owner:       owner,
		Name:        name,
		Description: repo.GetDescription(),
		URL:         repo.GetHTMLURL(),
		Language:    repo.GetLanguage(),
		Stars:       repo.GetStargazersCount(),
		IsOwner:     owner == username,
	}

	aiData := make(map[string]aiToolData)

	// Count commits by user
	commitOpts := &github.CommitsListOptions{
		Author:      username,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	if !opts.Since.IsZero() {
		commitOpts.Since = opts.Since
	}
	if !opts.Until.IsZero() {
		commitOpts.Until = opts.Until
	}

	for {
		commits, resp, err := c.gh.Repositories.ListCommits(ctx, owner, name, commitOpts)
		if err != nil {
			break // May not have access
		}
		contrib.Commits += len(commits)

		// Check each commit for AI co-authors
		for _, commit := range commits {
			commitDate := ""
			if commit.GetCommit().GetAuthor().GetDate() != (github.Timestamp{}) {
				commitDate = commit.GetCommit().GetAuthor().GetDate().Format("2006-01-02")
			}

			// Extract co-authors from commit message
			message := commit.GetCommit().GetMessage()
			matches := coAuthorRegex.FindAllStringSubmatch(message, -1)

			for _, match := range matches {
				if len(match) >= 3 {
					email := strings.ToLower(match[2])
					tool := GetAIToolByEmail(email)
					if tool != nil {
						data := aiData[tool.Name]
						data.Name = tool.Name
						data.Commits++
						data.Recognized = tool.Recognized
						if data.byDate == nil {
							data.byDate = make(map[string]int)
						}
						if commitDate != "" {
							data.byDate[commitDate]++
							if data.FirstUsed == "" || commitDate < data.FirstUsed {
								data.FirstUsed = commitDate
							}
							if commitDate > data.LastUsed {
								data.LastUsed = commitDate
							}
						}
						aiData[tool.Name] = data
					}
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		commitOpts.Page = resp.NextPage
	}

	// Count PRs by user
	prOpts := &github.PullRequestListOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		prs, resp, err := c.gh.PullRequests.List(ctx, owner, name, prOpts)
		if err != nil {
			break
		}

		for _, pr := range prs {
			if pr.GetUser().GetLogin() == username {
				createdAt := pr.GetCreatedAt().Time
				if !opts.Since.IsZero() && createdAt.Before(opts.Since) {
					continue
				}
				if !opts.Until.IsZero() && createdAt.After(opts.Until) {
					continue
				}
				contrib.PRs++
				if pr.GetMerged() {
					contrib.PRsMerged++
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		prOpts.Page = resp.NextPage
	}

	return contrib, aiData, nil
}

func (c *Client) getActivityHeatmap(ctx context.Context, opts ProfileOptions) ([]DailyActivity, error) {
	// Use GitHub's contribution calendar via GraphQL would be ideal,
	// but for now we'll aggregate from events
	activityMap := make(map[string]int)

	// Get user events
	eventOpts := &github.ListOptions{PerPage: 100}

	for page := 0; page < 10; page++ { // Limit to avoid rate limiting
		events, resp, err := c.gh.Activity.ListEventsPerformedByUser(ctx, opts.Username, false, eventOpts)
		if err != nil {
			break
		}

		for _, event := range events {
			createdAt := event.GetCreatedAt().Time
			if !opts.Since.IsZero() && createdAt.Before(opts.Since) {
				continue
			}
			if !opts.Until.IsZero() && createdAt.After(opts.Until) {
				continue
			}

			date := createdAt.Format("2006-01-02")
			activityMap[date]++
		}

		if resp.NextPage == 0 {
			break
		}
		eventOpts.Page = resp.NextPage
	}

	// Convert to slice
	var activity []DailyActivity
	for date, count := range activityMap {
		activity = append(activity, DailyActivity{Date: date, Count: count})
	}

	// Sort by date
	sort.Slice(activity, func(i, j int) bool {
		return activity[i].Date < activity[j].Date
	})

	return activity, nil
}
