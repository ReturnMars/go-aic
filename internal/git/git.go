package git

import (
	"fmt"
	"os/exec"
)

// GetDiff returns the staged changes (git diff --staged).
// If no staged changes, it returns error.
func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git error: %v", err)
	}
	return string(out), nil
}

// Commit executes git commit -m "message"
func Commit(message string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("commit failed: %v", err)
	}
	return string(out), nil
}

// CheckIfGitRepo checks if current directory is a git repo
func CheckIfGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}
