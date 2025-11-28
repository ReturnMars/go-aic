package tui

import (
	"fmt"
	"marsx/internal/ai"
	"marsx/internal/git"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SessionState int

const (
	StateStart     SessionState = iota // Check git repo
	StateInput                         // Waiting for user input (or auto-trigger)
	StateLoading                       // AI generating
	StateReview                        // Preview generated commit
	StateEditing                       // Editing commit message
	StateExecuting                     // Committing
	StateOutput                        // Done
	StateError
	StateChatResult
	StateConfirmAdd // New state for git add confirmation
)

type Model struct {
	viewport  viewport.Model
	textInput textinput.Model
	textArea  textarea.Model
	spinner   spinner.Model
	state     SessionState
	userInput string
	aiResult  string
	output    string
	err       error
	history   string
	aiClient  *ai.Client
	windowH   int
	diff      string
	quickMode bool
	FinalMsg  string // Exported for post-run display
}

func InitialModel(quick bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Type '?' to chat, or Enter to generate commit from staged files..."
	ti.Focus()
	ti.Width = 60

	ta := textarea.New()
	ta.Placeholder = "Commit message..."
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(highlight)

	vp := viewport.New(80, 1)
	initialContent := titleStyle.Render("MarsX Git Assistant") + "\n\nReady to commit.\n"
	vp.SetContent(initialContent)

	return Model{
		viewport:  vp,
		textInput: ti,
		textArea:  ta,
		spinner:   sp,
		state:     StateStart,
		history:   initialContent,
		aiClient:  ai.NewClient(),
		quickMode: quick,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, checkGitRepoCmd)
}

func (m *Model) resizeViewport() {
	const bottomHeight = 12
	maxHeight := m.windowH - bottomHeight
	if maxHeight < 1 {
		maxHeight = 1
	}
	contentHeight := lipgloss.Height(m.history)
	if contentHeight < maxHeight {
		m.viewport.Height = contentHeight
	} else {
		m.viewport.Height = maxHeight
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.state == StateEditing {
				m.state = StateReview
				m.textArea.Blur()
				return m, nil
			}
			if m.state == StateReview || m.state == StateConfirmAdd {
				m.state = StateInput
				m.textInput.Focus()
				return m, nil
			}
			if m.state == StateInput {
				return m, tea.Quit
			}
			m.state = StateInput
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink

		case tea.KeyEnter:
			if m.state == StateInput {
				m.userInput = m.textInput.Value()
				if strings.HasPrefix(m.userInput, "?") {
					m.state = StateLoading
					return m, tea.Batch(m.spinner.Tick, m.sendRequestCmd(m.userInput, ai.ModeChat))
				}
				m.state = StateLoading
				return m, tea.Batch(m.spinner.Tick, m.generateCommitCmd(m.userInput))

			} else if m.state == StateReview {
				msg := m.textArea.Value()
				if msg == "" {
					return m, nil
				}
				m.state = StateExecuting
				return m, tea.Batch(m.spinner.Tick, commitCmd(msg))
			} else if m.state == StateConfirmAdd {
				// Treat Enter as Yes
				return m, tea.Batch(m.spinner.Tick, gitAddAllCmd)
			}

		case tea.KeyCtrlS:
			if m.state == StateEditing {
				msg := m.textArea.Value()
				if msg == "" {
					return m, nil
				}
				m.state = StateExecuting
				return m, tea.Batch(m.spinner.Tick, commitCmd(msg))
			}
		}

		if m.state == StateReview && msg.String() == "e" {
			m.state = StateEditing
			m.textArea.Focus()
			return m, textarea.Blink
		}

		if m.state == StateConfirmAdd {
			if strings.ToLower(msg.String()) == "y" {
				return m, tea.Batch(m.spinner.Tick, gitAddAllCmd)
			} else if strings.ToLower(msg.String()) == "n" {
				m.state = StateInput
				m.textInput.Focus()
				return m, nil
			}
		}

		if m.state != StateInput && m.state != StateEditing && msg.String() == "q" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.windowH = msg.Height
		m.viewport.Width = msg.Width
		m.resizeViewport()

	case gitCheckMsg:
		if !msg.isRepo {
			m.err = fmt.Errorf("not a git repository")
			m.state = StateError
		} else {
			m.state = StateLoading
			return m, tea.Batch(m.spinner.Tick, m.generateCommitCmd(""))
		}

	case aiResponseMsg:
		if msg.mode == ai.ModeChat {
			m.state = StateChatResult
			newEntry := fmt.Sprintf("\n> %s\n%s\n", m.userInput, lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE")).Render(msg.content))
			m.history += newEntry
			m.viewport.SetContent(m.history)
			m.resizeViewport()
			m.viewport.GotoBottom()
			m.textInput.SetValue("")
		} else {
			m.state = StateReview
			m.aiResult = msg.content
			m.textArea.SetValue(msg.content)
			m.textArea.Blur()

			newEntry := fmt.Sprintf("\n> Generating commit...\nAI Suggestion: %s\n", msg.content)
			m.history += newEntry
			m.viewport.SetContent(m.history)
			m.resizeViewport()
			m.viewport.GotoBottom()
		}
		return m, nil

	case execOutputMsg:
		if strings.HasPrefix(msg.output, "Added") {
			// git add successful, now retry generation
			newEntry := fmt.Sprintf("\n> Staged all changes.\n")
			m.history += newEntry
			m.viewport.SetContent(m.history)
			m.resizeViewport()
			m.viewport.GotoBottom()

			m.state = StateLoading
			return m, tea.Batch(m.spinner.Tick, m.generateCommitCmd(""))
		}
		// Commit successful
		m.FinalMsg = fmt.Sprintf("✔ Commit successful!\n%s", msg.output)
		return m, tea.Quit

	case errMsg:
		if m.state == StateLoading && strings.Contains(msg.err.Error(), "no staged changes") {
			// Transition to ConfirmAdd state
			m.state = StateConfirmAdd
			return m, nil
		}

		m.state = StateError
		m.err = msg.err
		return m, nil

	case spinner.TickMsg:
		if m.state == StateLoading || m.state == StateExecuting {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if m.state == StateEditing {
		m.textArea, cmd = m.textArea.Update(msg)
		return m, cmd
	}

	if m.state == StateInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var bottomView string

	switch m.state {
	case StateStart:
		bottomView = "Checking git status..."
	case StateInput:
		bottomView = fmt.Sprintf(
			"%s\n%s",
			inputBoxStyle.Render(m.textInput.View()),
			helpStyle.Render("[Enter] Generate Commit • [? + text] Chat • [Esc] Quit"),
		)
	case StateLoading:
		bottomView = fmt.Sprintf("%s Processing...", m.spinner.View())
	case StateReview:
		bottomView = fmt.Sprintf(
			"Commit Message Preview:\n%s\n%s",
			commandStyle.Render(m.textArea.Value()),
			helpStyle.Render("[Enter] Commit • [e] Edit • [Esc] Cancel"),
		)
	case StateEditing:
		bottomView = fmt.Sprintf(
			"Editing Commit Message:\n%s\n%s",
			m.textArea.View(),
			helpStyle.Render("[Ctrl+S] Save & Commit • [Esc] Cancel Edit"),
		)
	case StateExecuting:
		bottomView = fmt.Sprintf("%s Executing...", m.spinner.View())
	case StateConfirmAdd:
		bottomView = fmt.Sprintf(
			"%s\n%s",
			lipgloss.NewStyle().Foreground(warning).Bold(true).Render("No staged changes found."),
			helpStyle.Render("Stage all changes (git add .)? [Y/n]"),
		)
	case StateOutput:
		bottomView = "Done."
	case StateError:
		bottomView = fmt.Sprintf(
			"Error: %s\n%s",
			lipgloss.NewStyle().Foreground(warning).Render(m.err.Error()),
			helpStyle.Render("[Esc] Back"),
		)
	case StateChatResult:
		bottomView = helpStyle.Render("Chat response received. Type next query.")
	}

	return lipgloss.JoinVertical(lipgloss.Left, viewportStyle.Render(m.viewport.View()), appStyle.Render(bottomView))
}

// ... (rest remains same)
type aiResponseMsg struct {
	content string
	mode    ai.Mode
}

type execOutputMsg struct {
	output string
}

type errMsg struct {
	err error
}

type gitCheckMsg struct {
	isRepo bool
}

func checkGitRepoCmd() tea.Msg {
	return gitCheckMsg{isRepo: git.CheckIfGitRepo()}
}

func (m Model) generateCommitCmd(hint string) tea.Cmd {
	return func() tea.Msg {
		diff, err := git.GetStagedDiff()
		if err != nil {
			return errMsg{err}
		}
		if len(diff) == 0 {
			return errMsg{fmt.Errorf("no staged changes")}
		}

		if len(diff) > 4000 {
			diff = diff[:4000] + "\n... (truncated)"
		}

		prompt := fmt.Sprintf("Diff:\n%s\n\nHint: %s", diff, hint)

		resp, err := m.aiClient.SendRequest(prompt, ai.ModeCommand)
		if err != nil {
			return errMsg{err}
		}
		return aiResponseMsg{content: resp, mode: ai.ModeCommand}
	}
}

func (m Model) sendRequestCmd(input string, mode ai.Mode) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.aiClient.SendRequest(input, mode)
		if err != nil {
			return errMsg{err}
		}
		return aiResponseMsg{content: resp, mode: mode}
	}
}

func commitCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		out, err := git.Commit(msg)
		if err != nil {
			return errMsg{err}
		}
		return execOutputMsg{output: out}
	}
}

func gitAddAllCmd() tea.Msg {
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return errMsg{fmt.Errorf("git add failed: %v", err)}
	}
	return execOutputMsg{output: "Added all changes"}
}
