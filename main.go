package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/octokit/go-sdk/pkg/github/models"
)

// appVersion is the version of the application.
const appVersion = "1.4.0"

// State represents the application state.
type State int

const (
	// StateInit is the initial state.
	StateInit State = iota
	// StateChecking is the state when checking if both releases exist.
	StateChecking
	// StateFetching is the state when fetching data from GitHub.
	StateFetching
	// StateDownloadExtract is the state when downloading and extracting all the releases.
	StateDownloadExtract
	// StateAnalyzing is the state when analyzing the downloaded releases.
	StateAnalyzing
	// StateSummary is the final state.
	StateSummary
)

var (
	ghRepo        = flag.String("repo", "", "GitHub repository to compare releases from. Format: owner/repo")
	ghToken       = flag.String("token", "", "GitHub token to use for API requests")
	firstRelease  = flag.String("from", "", "Base release to compare")
	secondRelease = flag.String("to", "", "Release to compare to")
	ignoreRegex   = flag.String("ignore", "", "Regex to ignore releases names from the analysis")
	extractionDir = flag.String("output", "releases", "Directory to extract releases to")
	remove        = flag.Bool(
		"remove", false,
		"Remove the directory containing the extracted releases once the processing is done",
	)
	version = flag.Bool("version", false, "Print the version and exit")

	docStyle    = lipgloss.NewStyle().Margin(1, 2)
	svelteColor = lipgloss.Color("#ff3e00")
	svelteText  = lipgloss.NewStyle().Foreground(svelteColor)
	svelteBg    = lipgloss.NewStyle().Background(svelteColor).Foreground(
		lipgloss.AdaptiveColor{
			Light: "#ffffff",
			Dark:  "#000000",
		},
	)
	blurredSvelteText = lipgloss.NewStyle().Foreground(lipgloss.Color("#cc5833"))
	blurredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	noStyle           = lipgloss.NewStyle()
)

type (
	// fatalErr is a fatal error message.
	fatalErr struct{}

	// data is the application data model.
	data struct {
		ghRepo        string               // GitHub repository to compare releases from. Format: owner/repo
		ghToken       string               // GitHub token to use for API requests
		firstRelease  string               // Base release to compare
		secondRelease string               // Release to compare to
		ignoreRegex   string               // Regex to ignore releases names from the analysis
		releases      []models.Releaseable // GitHub releases
		analysis      []AnalysisResult     // Analysis results
	}
)

func initialModel() model {
	flag.Parse()

	// Print version and exit
	if *version {
		exe, err := os.Executable()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(filepath.Base(exe), appVersion)
		os.Exit(0)
	}

	m := model{
		data: data{
			ghRepo:        *ghRepo,
			ghToken:       *ghToken,
			firstRelease:  *firstRelease,
			secondRelease: *secondRelease,
			ignoreRegex:   *ignoreRegex,
		},
		tarSize: make(map[string]int64),
	}

	// Initialize spinner
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = svelteText
	m.spinner = spin

	// Initialize text inputs
	if m.data.ghRepo == "" {
		input := textinput.New()
		input.Placeholder = "GitHub repository (owner/repo)"
		m.inputs = append(m.inputs, input)
	}
	if m.data.ghToken == "" {
		tokenInput := textinput.New()
		tokenInput.Placeholder = "GitHub token (optional)"
		tokenInput.EchoMode = textinput.EchoPassword
		tokenInput.EchoCharacter = '•'
		m.inputs = append(m.inputs, tokenInput)
	}
	if m.data.firstRelease == "" {
		input := textinput.New()
		input.Placeholder = "Base release"
		m.inputs = append(m.inputs, input)
	}
	if m.data.secondRelease == "" {
		input := textinput.New()
		input.Placeholder = "Release to compare to"
		m.inputs = append(m.inputs, input)
	}
	if m.data.ignoreRegex == "" {
		input := textinput.New()
		input.Placeholder = "Regex to ignore releases names (optional)"
		m.inputs = append(m.inputs, input)
	}

	// Focus the first input
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
		m.inputs[0].Cursor.Style = svelteText
		m.inputs[0].PromptStyle = svelteText
	}

	return m
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error running program:", err)
		os.Exit(1)
	}
}
