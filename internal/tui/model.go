package tui

import (
	"fmt"
	"marsx/internal/ai"
	"marsx/internal/git"
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
	StateReview                        // User reviewing/editing commit message
	StateExecuting                     // Committing
	StateOutput                        // Done
	StateError
	StateChatResult
)

type Model struct {
	viewport  viewport.Model
	textInput textinput.Model // For chat/supplementary input
	textArea  textarea.Model  // For editing commit message
	spinner   spinner.Model
	state     SessionState
	userInput string
	aiResult  string // The generated commit message
	output    string
	err       error
	history   string
	aiClient  *ai.Client
	windowH   int
	diff      string // Staged diff
}

func InitialModel() Model {
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
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, checkGitRepoCmd)
}

func (m *Model) resizeViewport() {
	const bottomHeight = 12 // Increased for TextArea
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
			if m.state == StateReview {
				// Cancel review, back to input
				m.state = StateInput
				m.textInput.Focus()
				m.textArea.Blur()
				return m, nil
			}
			if m.state == StateInput {
				return m, tea.Quit
			}
			// Reset
			m.state = StateInput
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink

		case tea.KeyEnter:
			if m.state == StateInput {
				m.userInput = m.textInput.Value()

				// Chat mode
				if strings.HasPrefix(m.userInput, "?") {
					m.state = StateLoading
					return m, tea.Batch(m.spinner.Tick, m.sendRequestCmd(m.userInput, ai.ModeChat))
				}

				// Commit Generation Mode
				// If user typed something, treat it as context hint. If empty, just use diff.
				m.state = StateLoading
				return m, tea.Batch(m.spinner.Tick, m.generateCommitCmd(m.userInput))

			} else if m.state == StateReview {
				// User confirmed commit message
				msg := m.textArea.Value()
				if msg == "" {
					return m, nil // Don't commit empty
				}
				m.state = StateExecuting
				return m, tea.Batch(m.spinner.Tick, commitCmd(msg))
			}
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
			m.state = StateInput
		}

	case aiResponseMsg:
		if msg.mode == ai.ModeChat {
			m.state = StateChatResult
			newEntry := fmt.Sprintf("\n> %s\n%s\n", m.userInput, lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE")).Render(msg.content))
			m.history += newEntry
			m.viewport.SetContent(m.history)
			m.resizeViewport()
			m.viewport.GotoBottom()
			// Reset input
			m.textInput.SetValue("")
		} else {
			// Commit generated
			m.state = StateReview
			m.aiResult = msg.content
			m.textArea.SetValue(msg.content)
			m.textArea.Focus()
			m.textInput.Blur() // Blur main input

			// Show diff summary in history
			newEntry := fmt.Sprintf("\n> Generating commit for staged changes...\nAI Suggestion: %s\n", msg.content)
			m.history += newEntry
			m.viewport.SetContent(m.history)
			m.resizeViewport()
			m.viewport.GotoBottom()
		}
		return m, nil

	case execOutputMsg:
		newEntry := fmt.Sprintf("\n> Commit successful!\n%s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa")).Render(msg.output))
		m.history += newEntry
		m.viewport.SetContent(m.history)
		m.resizeViewport()
		m.viewport.GotoBottom()

		m.state = StateOutput
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.textArea.Reset()
		return m, nil

	case errMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case spinner.TickMsg:
		if m.state == StateLoading || m.state == StateExecuting {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if m.state == StateReview {
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
			helpStyle.Render("[Enter] Generate Commit (staged) • [? + text] Chat • [Esc] Quit"),
		)
	case StateLoading:
		bottomView = fmt.Sprintf("%s Processing...", m.spinner.View())
	case StateReview:
		bottomView = fmt.Sprintf(
			"Edit Commit Message:\n%s\n%s",
			m.textArea.View(),
			helpStyle.Render("[Enter] Commit • [Esc] Cancel"),
		)
	case StateExecuting:
		bottomView = fmt.Sprintf("%s Committing...", m.spinner.View())
	case StateOutput:
		bottomView = fmt.Sprintf(
			"%s\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(special).Render("Done!"),
			helpStyle.Render("[Enter] New Action • [q] Quit"),
		)
	case StateError:
		bottomView = fmt.Sprintf(
			"Error: %s\n%s",
			lipgloss.NewStyle().Foreground(warning).Render(m.err.Error()),
			helpStyle.Render("[Esc] Back"),
		)
	case StateChatResult:
		bottomView = helpStyle.Render("Chat response received. Type next query or command.")
	}

	return lipgloss.JoinVertical(lipgloss.Left, viewportStyle.Render(m.viewport.View()), appStyle.Render(bottomView))
}

// Messages & Cmds

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
			return errMsg{fmt.Errorf("no staged changes. Use 'git add' first")}
		}

		// Limit diff size to avoid token limits
		if len(diff) > 4000 {
			diff = diff[:4000] + "\n... (truncated)"
		}

		prompt := fmt.Sprintf("Diff:\n%s\n\nContext hint: %s", diff, hint)

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
