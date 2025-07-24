package client

import "os"

// isGitHubActions returns true if running in GitHub Actions
func isGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") != ""
}
