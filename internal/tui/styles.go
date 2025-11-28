package tui

import "github.com/charmbracelet/lipgloss"

var (
	// 颜色定义 (Charm 风格)
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	warning   = lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F5508B"}

	// 基础样式
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFdf5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(0, 1).
			Width(60)

	commandStyle = lipgloss.NewStyle().
			Foreground(special).
			Background(lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#2B2B2B"}).
			Padding(1, 2).
			Margin(1, 0).
			Bold(true)

	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(subtle).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtle).
			MarginTop(1)
)
