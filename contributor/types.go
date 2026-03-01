package contributor

import (
	"strings"
	"time"
)

// Profile represents an individual contributor's development profile.
type Profile struct {
	Username     string            `json:"username"`
	Name         string            `json:"name,omitempty"`
	AvatarURL    string            `json:"avatarUrl,omitempty"`
	Bio          string            `json:"bio,omitempty"`
	Company      string            `json:"company,omitempty"`
	Location     string            `json:"location,omitempty"`
	Blog         string            `json:"blog,omitempty"`
	Repositories []RepoContrib     `json:"repositories"`
	Stats        ContributorStats  `json:"stats"`
	AIStats      AICollabStats     `json:"aiStats"`
	Activity     []DailyActivity   `json:"activity"`
	Languages    map[string]int    `json:"languages"`
	DateRange    DateRange         `json:"dateRange"`
	GeneratedAt  time.Time         `json:"generatedAt"`
}

// RepoContrib represents contributions to a single repository.
type RepoContrib struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	Commits     int    `json:"commits"`
	PRs         int    `json:"prs"`
	PRsMerged   int    `json:"prsMerged"`
	Issues      int    `json:"issues"`
	Reviews     int    `json:"reviews"`
	Language    string `json:"language,omitempty"`
	Stars       int    `json:"stars"`
	IsOwner     bool   `json:"isOwner"`
}

// ContributorStats holds aggregate statistics for a contributor.
type ContributorStats struct {
	TotalCommits      int `json:"totalCommits"`
	TotalPRs          int `json:"totalPRs"`
	TotalPRsMerged    int `json:"totalPRsMerged"`
	TotalIssues       int `json:"totalIssues"`
	TotalReviews      int `json:"totalReviews"`
	TotalRepositories int `json:"totalRepositories"`
	OwnedRepos        int `json:"ownedRepos"`
	ContributedRepos  int `json:"contributedRepos"`
}

// DailyActivity represents activity for a single day (for heatmap).
type DailyActivity struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"` // Total contributions
}

// DateRange specifies the date range for the profile.
type DateRange struct {
	Start string `json:"start"` // YYYY-MM-DD
	End   string `json:"end"`   // YYYY-MM-DD
}

// ProgressFunc is called to report progress during profile generation.
// stage is 1-based, totalStages is the total number of stages.
// current/total are for progress within the current stage (0 if not applicable).
// description describes what's happening.
type ProgressFunc func(stage, totalStages int, current, total int, description string, done bool)

// ProfileOptions configures profile generation.
type ProfileOptions struct {
	Username   string
	Orgs       []string     // Filter to specific organizations
	Since      time.Time    // Start date
	Until      time.Time    // End date
	LocalPaths []string     // Local paths to search for repos (e.g., ~/go/src)
	APIOnly    bool         // Force API-only mode, skip local repo detection
	Progress   ProgressFunc // Progress callback (optional)
}

// ContributionEvent represents a single contribution event.
type ContributionEvent struct {
	Type      string    `json:"type"` // commit, pr, issue, review
	Repo      string    `json:"repo"`
	Title     string    `json:"title,omitempty"`
	URL       string    `json:"url,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// AICollabStats tracks AI-assisted development metrics.
// This measures how "AI-native" a developer is by tracking
// commits with AI tool co-authors.
type AICollabStats struct {
	TotalAICommits   int                   `json:"totalAiCommits"`
	AICommitPercent  float64               `json:"aiCommitPercent"`
	ByTool           map[string]AIToolStat `json:"byTool"`
	MostUsedTool     string                `json:"mostUsedTool,omitempty"`
	FirstAICommit    string                `json:"firstAiCommit,omitempty"`    // Date of first AI-assisted commit
	RecentTrend      string                `json:"recentTrend,omitempty"`      // increasing, stable, decreasing
	AIActivity       []DailyActivity       `json:"aiActivity,omitempty"`       // Daily AI-assisted commits for heatmap
}

// AIToolStat holds statistics for a specific AI coding tool.
type AIToolStat struct {
	Name        string `json:"name"`
	Commits     int    `json:"commits"`
	FirstUsed   string `json:"firstUsed,omitempty"`   // YYYY-MM-DD
	LastUsed    string `json:"lastUsed,omitempty"`    // YYYY-MM-DD
	Recognized  bool   `json:"recognized"`            // true if GitHub recognizes the co-author
}

// KnownAITool represents a known AI coding assistant.
type KnownAITool struct {
	Name       string   // Display name
	Emails     []string // Known email addresses
	Recognized bool     // Whether GitHub recognizes it as a user
}

// KnownAITools lists recognized AI coding assistants and their co-author signatures.
var KnownAITools = []KnownAITool{
	{
		Name: "Claude Code",
		Emails: []string{
			"noreply@anthropic.com",
		},
		Recognized: true, // GitHub recognizes this email
	},
	{
		Name: "GitHub Copilot",
		Emails: []string{
			"noreply@github.com",
			"copilot@github.com",
		},
		Recognized: true,
	},
	{
		Name: "Gemini CLI",
		Emails: []string{
			// Official recommended format: Co-authored-by: gemini-cli ${MODEL} <218195315+gemini-cli@users.noreply.github.com>
			"218195315+gemini-cli@users.noreply.github.com",
			// Gemini Code Assist bot
			"176961590+gemini-code-assist[bot]@users.noreply.github.com",
			// Community-used variants
			"gemini-cli-agent@google.com",
			"gemini@google.com",
		},
		Recognized: true, // The official noreply format IS recognized by GitHub
	},
	{
		Name: "Cursor",
		Emails: []string{
			"ai@cursor.sh",
			"cursor@cursor.sh",
		},
		Recognized: false, // Verify actual email pattern
	},
	{
		Name: "Aider",
		Emails: []string{
			"aider@aider.chat",
		},
		Recognized: false,
	},
}

// GetAIToolByEmail returns the AI tool matching the given email, if any.
// Handles both exact matches and suffix matches for GitHub noreply formats
// like "123456+username@users.noreply.github.com".
func GetAIToolByEmail(email string) *KnownAITool {
	email = strings.ToLower(email)

	for i := range KnownAITools {
		for _, pattern := range KnownAITools[i].Emails {
			pattern = strings.ToLower(pattern)

			// Exact match
			if email == pattern {
				return &KnownAITools[i]
			}

			// Suffix match for GitHub noreply with variable user ID prefix
			// e.g., "218195315+gemini-cli@users.noreply.github.com" matches
			// any email ending in "+gemini-cli@users.noreply.github.com"
			if strings.Contains(pattern, "+") && strings.HasSuffix(pattern, "@users.noreply.github.com") {
				// Extract the part after the + sign
				plusIdx := strings.Index(pattern, "+")
				suffix := pattern[plusIdx:]
				if strings.HasSuffix(email, suffix) {
					return &KnownAITools[i]
				}
			}
		}
	}
	return nil
}
