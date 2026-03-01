package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/mogo/fmt/progress"
	"github.com/spf13/cobra"

	"github.com/plexusone/devfolio/contributor"
)

var (
	contribProfileUser      string
	contribProfileOutput    string
	contribProfileSince     string
	contribProfileUntil     string
	contribProfileOrgs      []string
	contribProfileAPIOnly   bool
	contribProfileLocalPath string
)

var contributorProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Generate contributor profile/portfolio",
	Long: `Generate a comprehensive contributor profile from GitHub activity.

The profile includes:
  - Contribution summary (commits, PRs, issues, reviews)
  - Activity heatmap (GitHub-style contribution calendar)
  - Repository breakdown
  - Language statistics
  - Timeline of notable contributions

Requires GITHUB_TOKEN environment variable for API access.

Examples:
  # Generate profile for a user
  devfolio contributor profile --user grokify -o profile.json

  # Limit to specific organizations
  devfolio contributor profile --user grokify --org fleet-ops --org agentplexus

  # Filter by date range
  devfolio contributor profile --user grokify --since 2024-01-01`,
	RunE: runContributorProfile,
}

func init() {
	contributorProfileCmd.Flags().StringVar(&contribProfileUser, "user", "", "GitHub username (required)")
	contributorProfileCmd.Flags().StringVarP(&contribProfileOutput, "output", "o", "", "Output file (default: stdout)")
	contributorProfileCmd.Flags().StringVar(&contribProfileSince, "since", "", "Start date (YYYY-MM-DD)")
	contributorProfileCmd.Flags().StringVar(&contribProfileUntil, "until", "", "End date (YYYY-MM-DD)")
	contributorProfileCmd.Flags().StringArrayVar(&contribProfileOrgs, "org", nil, "Filter to specific organizations")
	contributorProfileCmd.Flags().BoolVar(&contribProfileAPIOnly, "api-only", false, "Force API-only mode, skip local repo detection")
	contributorProfileCmd.Flags().StringVar(&contribProfileLocalPath, "local-path", "", "Additional local path to search for repos")
	_ = contributorProfileCmd.MarkFlagRequired("user")
	contributorCmd.AddCommand(contributorProfileCmd)
}

func runContributorProfile(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	// Set up progress renderer
	renderer := progress.NewMultiStageRenderer(os.Stderr)

	// Build local paths list
	var localPaths []string
	home, _ := os.UserHomeDir()
	if home != "" {
		// Default search paths
		localPaths = append(localPaths, filepath.Join(home, "go", "src", "github.com"))
	}
	if contribProfileLocalPath != "" {
		localPaths = append(localPaths, contribProfileLocalPath)
	}

	// Build profile options
	opts := contributor.ProfileOptions{
		Username:   contribProfileUser,
		Orgs:       contribProfileOrgs,
		APIOnly:    contribProfileAPIOnly,
		LocalPaths: localPaths,
		Progress: func(stage, totalStages, current, total int, description string, done bool) {
			renderer.Update(progress.StageInfo{
				Stage:       stage,
				TotalStages: totalStages,
				Current:     current,
				Total:       total,
				Description: description,
				Done:        done,
			})
		},
	}

	if contribProfileSince != "" {
		t, err := time.Parse("2006-01-02", contribProfileSince)
		if err != nil {
			return fmt.Errorf("invalid --since date: %w", err)
		}
		opts.Since = t
	}

	if contribProfileUntil != "" {
		t, err := time.Parse("2006-01-02", contribProfileUntil)
		if err != nil {
			return fmt.Errorf("invalid --until date: %w", err)
		}
		opts.Until = t
	}

	// Generate profile
	client, err := contributor.NewClient(token)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	profile, err := client.GenerateProfile(ctx, opts)
	if err != nil {
		return fmt.Errorf("generating profile: %w", err)
	}

	// Output
	output, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling profile: %w", err)
	}

	if contribProfileOutput == "" {
		fmt.Println(string(output))
	} else {
		if err := os.WriteFile(contribProfileOutput, output, 0600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote contributor profile to %s\n", contribProfileOutput)
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\nContributor profile summary:\n")
	fmt.Fprintf(os.Stderr, "  Username:     %s\n", profile.Username)
	fmt.Fprintf(os.Stderr, "  Repositories: %d\n", len(profile.Repositories))
	fmt.Fprintf(os.Stderr, "  Commits:      %d\n", profile.Stats.TotalCommits)
	fmt.Fprintf(os.Stderr, "  PRs:          %d\n", profile.Stats.TotalPRs)
	if profile.AIStats.TotalAICommits > 0 {
		fmt.Fprintf(os.Stderr, "  AI Commits:   %d (%.1f%%)\n", profile.AIStats.TotalAICommits, profile.AIStats.AICommitPercent)
	}

	return nil
}
