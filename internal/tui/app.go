package tui

import (
	"bufio"
	"fmt"
	"marsx/internal/ai"
	"marsx/internal/git"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func Start(quickMode bool, chatMode bool) {
	fmt.Println(colorGreen + colorBold + "\n🚀 MarsX Git Assistant" + colorReset)
	fmt.Println(colorGray + "----------------------------------------" + colorReset)

	if !git.CheckIfGitRepo() {
		fmt.Println(colorRed + "Error: Not a git repository." + colorReset)
		return
	}

	client := ai.NewClient()
	reader := bufio.NewReader(os.Stdin)
	var diff string
	var err error

	// Only check diff if NOT in chat mode
	if !chatMode {
		// Auto-check stage status
		diff, err = git.GetStagedDiff()
		// Check for error OR empty diff (git command success but empty output)
		if err != nil || diff == "" {
			// Likely no staged changes
			fmt.Println(colorYellow + "⚠ No staged changes found." + colorReset)
			fmt.Print(colorBold + "Press [Enter] to stage all changes (git add .), or type 'n' to skip: " + colorReset)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			// Enter (empty string) or 'y' means YES
			if input == "" || input == "y" {
				if err := git.AddAll(); err != nil {
					fmt.Printf(colorRed+"Error adding files: %v\n"+colorReset, err)
					return
				}
				fmt.Println(colorGreen + "✔ All changes staged." + colorReset)
				// Update diff
				diff, _ = git.GetStagedDiff()
			}
		}
	} else {
		fmt.Println(colorYellow + "ℹ Chat Mode Enabled" + colorReset)
	}

	// If quick mode or we have a diff, try to generate immediately
	// But skip if chatMode is explicitly set
	var currentCommitMsg string
	if !chatMode && (quickMode || diff != "") && diff != "" {
		currentCommitMsg = generateCommit(client, diff)
	}

	// Main Loop
	for {
		if currentCommitMsg != "" {
			fmt.Println(colorGray + "\nGenerated Commit Message:" + colorReset)
			fmt.Println(colorCyan + "----------------------------------------" + colorReset)
			fmt.Println(colorBold + currentCommitMsg + colorReset)
			fmt.Println(colorCyan + "----------------------------------------" + colorReset)
			fmt.Println(colorGray + "[Enter] Commit  [e] Edit  [c] Chat  [q] Quit" + colorReset)
		} else {
			fmt.Println(colorGray + "\n[Type message] Chat  [Enter] Generate from Diff  [q] Quit" + colorReset)
		}

		fmt.Print(colorGreen + "> " + colorReset)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			fmt.Println("Bye! 👋")
			break
		}

		// Logic branch
		if currentCommitMsg != "" {
			// We have a pending commit message
			if input == "" {
				// Commit
				out, err := git.Commit(currentCommitMsg)
				if err != nil {
					fmt.Printf(colorRed+"Commit failed: %v\n"+colorReset, err)
				} else {
					fmt.Println(colorGreen + "✔ Commit successful!" + colorReset)
					fmt.Println(out)
					break // Exit after successful commit
				}
			} else if input == "e" {
				// Edit using git commit -e -m
				fmt.Println(colorYellow + "Opening editor..." + colorReset)
				cmd := exec.Command("git", "commit", "-e", "-m", currentCommitMsg)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf(colorRed+"Commit cancelled or failed: %v\n"+colorReset, err)
				} else {
					fmt.Println(colorGreen + "✔ Commit successful!" + colorReset)
					break
				}
			} else if input == "c" {
				// Switch to chat mode, clear current commit msg
				currentCommitMsg = ""
				fmt.Println(colorYellow + "Switched to Chat Mode." + colorReset)
				continue
			} else {
				// Treat as refinement
				fmt.Println(colorBlue + "Refining with AI..." + colorReset)
				resp := chatWithAI(client, fmt.Sprintf("Previous message: %s\nUser request: %s\nPlease provide a revised commit message based on the request.", currentCommitMsg, input), ai.ModeCommand)
				currentCommitMsg = resp
			}
		} else {
			// No pending commit message
			if input == "" {
				// Trigger generation
				diff, err = git.GetStagedDiff()
				if err != nil || diff == "" {
					fmt.Println(colorRed + "No staged changes to generate from." + colorReset)
				} else {
					currentCommitMsg = generateCommit(client, diff)
				}
			} else {
				// Chat
				_ = chatWithAI(client, input, ai.ModeChat)
				// Print response (streaming effect is inside chatWithAI)
				// Just ensure a newline at the end
				fmt.Println()
			}
		}
	}
}

func generateCommit(client *ai.Client, diff string) string {
	if len(diff) > 4000 {
		diff = diff[:4000] + "\n... (truncated)"
	}
	fmt.Print(colorBlue + "Thinking..." + colorReset)

	prompt := fmt.Sprintf("Diff:\n%s", diff)
	msg := chatWithAI(client, prompt, ai.ModeCommand)
	// Clear the "Thinking..." line if possible, or just print newline
	fmt.Print("\r" + strings.Repeat(" ", 20) + "\r")
	return msg
}

func chatWithAI(client *ai.Client, input string, mode ai.Mode) string {
	// Start spinner
	done := make(chan bool)
	go func() {
		chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r"+colorCyan+"%s AI is processing..."+colorReset, chars[i%len(chars)])
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()

	resp, err := client.SendRequest(input, mode)
	done <- true
	fmt.Print("\r\033[K") // Clear line

	if err != nil {
		fmt.Printf(colorRed+"Error: %v\n"+colorReset, err)
		return ""
	}

	// Use Glamour to render markdown for Chat responses
	if mode == ai.ModeChat {
		r, _ := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(80),
		)
		out, err := r.Render(resp)
		if err == nil {
			fmt.Print(out)
			return resp
		}
		// Fallback to plain text if render fails
	}

	// Fallback / Command mode: print plain text (maybe typewriter effect?)
	// For command mode, we usually want raw text for editing/copying, so no markdown.
	if mode == ai.ModeCommand {
		return resp
	}

	// If we are here, it means either chat mode markdown failed or we want simple output
	for _, char := range resp {
		fmt.Print(string(char))
		time.Sleep(10 * time.Millisecond)
	}
	return resp
}
