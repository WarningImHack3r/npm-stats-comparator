package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
)

// model is the application internal state.
type model struct {
	data  data
	state State

	spinner spinner.Model

	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode

	existingReleasesCount uint

	downloadProgress   uint
	downloadCacheCount uint
	tarSize            map[string]int64

	list                      *list.Model
	wantedWidth, wantedHeight *int

	err error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return m
		},
		m.spinner.Tick,
	)
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case fatalErr:
		time.Sleep(250 * time.Millisecond) // Wait for the view to render
		os.Exit(1)
	case model:
		if m.state == StateInit && len(m.inputs) == 0 {
			m.state++ // Move to StateChecking
			_, spinCmd := m.spinner.Update(msg)
			return m, tea.Batch(
				spinCmd,
				DoesGitHubReleaseExist(m.data.ghRepo, m.data.ghToken, m.data.firstRelease),
				DoesGitHubReleaseExist(m.data.ghRepo, m.data.ghToken, m.data.secondRelease),
			)
		}
	case tea.KeyMsg:
		switch typ := msg.Type; typ {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.list != nil && m.list.FilterState() == list.Filtering && typ != tea.KeyCtrlC {
				break
			}
			// Quit
			return m, tea.Quit
		case tea.KeyCtrlR:
			if m.state != StateInit {
				break
			}
			// Change cursor mode
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			commands := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				commands[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(commands...)
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			if m.state != StateInit {
				break
			}
			// Did the user press enter while the "submit" button was focused?
			if typ == tea.KeyEnter && m.focusIndex == len(m.inputs) {
				// Get back the info from the inputs
				inputIndex := 0
				if m.data.ghRepo == "" {
					m.data.ghRepo = m.inputs[inputIndex].Value()
					if m.data.ghRepo == "" || strings.Count(m.data.ghRepo, "/") != 1 {
						// Invalid GitHub repository format
						m.err = fmt.Errorf("invalid GitHub repository format. Format: owner/repo")
						break
					}
					inputIndex++
				}
				if m.data.ghToken == "" {
					m.data.ghToken = m.inputs[inputIndex].Value()
					inputIndex++
				}
				if m.data.firstRelease == "" {
					m.data.firstRelease = m.inputs[inputIndex].Value()
					if m.data.firstRelease == "" {
						// Invalid first release
						m.err = fmt.Errorf("invalid base release")
						break
					}
					inputIndex++
				}
				if m.data.secondRelease == "" {
					m.data.secondRelease = m.inputs[inputIndex].Value()
					if m.data.secondRelease == "" {
						// Invalid second release
						m.err = fmt.Errorf("invalid release to compare to")
						break
					}
					inputIndex++
				}
				if m.data.ignoreRegex == "" {
					m.data.ignoreRegex = m.inputs[inputIndex].Value()
				}

				m.state++ // Move to StateChecking
				return m, tea.Batch(
					DoesGitHubReleaseExist(m.data.ghRepo, m.data.ghToken, m.data.firstRelease),
					DoesGitHubReleaseExist(m.data.ghRepo, m.data.ghToken, m.data.secondRelease),
				)
			}

			// Cycle indexes
			if typ == tea.KeyUp || typ == tea.KeyShiftTab {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			commands := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					commands[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = svelteText
					m.inputs[i].Cursor.Style = svelteText
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].Cursor.Style = noStyle
			}

			return m, tea.Batch(commands...)
		default:
			if m.state != StateInit {
				break
			}
			return m, func() tea.Cmd {
				// Update all inputs
				commands := make([]tea.Cmd, len(m.inputs))

				// Only text inputs with Focus() set will respond, so it's safe to simply
				// update all of them here without any further logic.
				for i := range m.inputs {
					m.inputs[i], commands[i] = m.inputs[i].Update(msg)
				}

				return tea.Batch(commands...)
			}()
		}
	case errMsg:
		m.err = msg
	case gitReleaseExistsMsg:
		if msg.exists {
			m.existingReleasesCount++
			if m.existingReleasesCount == 2 {
				m.state++ // Move to StateFetching
				_, spinCmd := m.spinner.Update(msg)
				return m, tea.Batch(
					spinCmd,
					GetGitHubReleases(
						m.data.ghRepo,
						m.data.ghToken,
						m.data.firstRelease,
						m.data.secondRelease,
						m.data.ignoreRegex,
					),
				)
			}
		} else {
			m.err = fmt.Errorf(
				"%s does not exist, check that you input an existing GitHub tag"+
					" (check at https://github.com/%s/tags)", msg.release, m.data.ghRepo,
			)
		}
	case gitReleasesDownloadSuccessMsg:
		m.data.releases = msg
		m.state++ // Move to StateDownloadExtract
		if len(m.data.releases) == 0 {
			m.err = fmt.Errorf("no releases found, please check your inputs")
			break
		}
		_, spinCmd := m.spinner.Update(msg)
		commands := make([]tea.Cmd, len(m.data.releases)+1)
		commands[0] = spinCmd
		for i, release := range m.data.releases {
			tagName := release.GetTagName()
			if tagName == nil {
				continue
			}
			commands[i+1] = DownloadGitHubRelease(
				*tagName, *extractionDir,
			)
		}
		return m, tea.Batch(commands...)
	case gitReleaseDownloadedMsg:
		m.downloadProgress++
		if msg.cached {
			m.downloadCacheCount++
		} else {
			m.tarSize[msg.release] = msg.tarSize
		}
		if m.downloadProgress == uint(len(m.data.releases)) {
			m.state++ // Move to StateAnalyzing
			_, spinCmd := m.spinner.Update(msg)
			analysis := make([]tea.Cmd, len(m.data.releases)+1)
			analysis[0] = spinCmd
			for i, release := range m.data.releases {
				tagName := release.GetTagName()
				if tagName == nil {
					continue
				}
				analysis[i+1] = AnalyzeRelease(*extractionDir, *tagName)
			}
			return m, tea.Batch(analysis...)
		}
	case analysisDoneMsg:
		// Initialize the analysis slice if it's empty
		if len(m.data.analysis) == 0 {
			m.data.analysis = make([]AnalysisResult, len(m.data.releases))
		}
		// Get index of the release in m.data.releases
		index := -1
		for i, release := range m.data.releases {
			tagName := release.GetTagName()
			if tagName == nil {
				continue
			}
			if *tagName == msg.releaseTag {
				index = i
				break
			}
		}
		if index == -1 {
			break
		}
		m.data.analysis[index] = msg // Insert the analysis result
		if v, ok := m.tarSize[msg.releaseTag]; ok {
			m.data.analysis[index].tarSize = v
		}

		areAllAnalysesDone := true
		for _, analysis := range m.data.analysis {
			if analysis.releaseTag == "" {
				areAllAnalysesDone = false
				break
			}
		}
		if areAllAnalysesDone {
			// Remove the directory containing the extracted releases
			if *remove {
				if err := os.RemoveAll(*extractionDir); err != nil {
					m.err = err
					break
				}
			}

			// Populate the list
			items := make([]ListItem, len(m.data.analysis))
			for i, analysis := range m.data.analysis {
				item := ListItem{AnalysisResult: analysis}
				if i > 0 {
					item.next = &items[i-1]
				}
				items[i] = item
			}
			for i := len(items) - 1; i >= 0; i-- {
				if i < len(items)-1 {
					items[i].previous = &items[i+1]
				}
			}
			listItems := make([]list.Item, len(items))
			for i, item := range items {
				listItems[i] = item
			}

			// Create the list
			l := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
			l.Title = "Releases comparison"
			l.Styles.Title = svelteBg.Padding(0, 1)
			l.Styles.FilterPrompt = svelteText
			l.Styles.FilterCursor = svelteText // FIXME: Those two styles don't seem to work
			m.list = &l
			if m.wantedWidth != nil && m.wantedHeight != nil {
				m.list.SetSize(*m.wantedWidth, *m.wantedHeight)
			}

			m.state++ // Move to StateSummary
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		if m.list != nil {
			m.wantedWidth, m.wantedHeight = nil, nil
			m.list.SetSize(msg.Width-h, msg.Height-v)
		} else {
			wantedWidth, wantedHeight := msg.Width-h, msg.Height-v
			m.wantedWidth, m.wantedHeight = &wantedWidth, &wantedHeight
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.list != nil {
		listModel, cmd := m.list.Update(message)
		m.list = &listModel
		return m, cmd
	}

	if m.err != nil {
		return m, func() tea.Msg {
			return fatalErr{}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	var builder strings.Builder

	switch m.state {
	case StateInit:
		builder.WriteRune('\n')
		for i := range m.inputs {
			if i > 0 {
				builder.WriteRune('\n')
			}
			builder.WriteString(m.inputs[i].View())
		}

		button := "[ Submit ]"
		if m.focusIndex == len(m.inputs) {
			button = svelteText.Render(button)
		}
		_, err := fmt.Fprintf(&builder, "\n\n%s\n\n", button)
		if err != nil {
			return ""
		}

		builder.WriteString(blurredStyle.Render("cursor mode is "))
		builder.WriteString(blurredSvelteText.Render(m.cursorMode.String()))
		builder.WriteString(blurredStyle.Render(fmt.Sprintf(" (%s to change style)", tea.KeyCtrlR.String())))
	case StateChecking:
		if m.existingReleasesCount < 2 {
			builder.WriteString(fmt.Sprintf("\n   %s Checking if releases exist...\n", m.spinner.View()))
		}
	case StateFetching:
		if m.data.releases == nil {
			builder.WriteString(fmt.Sprintf("\n   %s Fetching releases...\n", m.spinner.View()))
		}
	case StateDownloadExtract:
		builder.WriteString(
			fmt.Sprintf(
				"\n   %s Downloading and extracting releases (%d/%d",
				m.spinner.View(),
				m.downloadProgress,
				len(m.data.releases),
			),
		)
		if m.downloadCacheCount > 0 {
			builder.WriteString(fmt.Sprintf(" - %d cached", m.downloadCacheCount))
		}
		builder.WriteString(")...\n")
		builder.WriteString(
			blurredStyle.Render(
				fmt.Sprintf("     Downloaded versions are available in the `%s/` directory", *extractionDir),
			),
		)
	case StateAnalyzing:
		builder.WriteString(
			fmt.Sprintf(
				"\n   %s Analyzing releases (%d/%d)...\n",
				m.spinner.View(),
				len(m.data.analysis),
				len(m.data.releases),
			),
		)
	case StateSummary:
		builder.WriteString(docStyle.Render(m.list.View()))
	}

	return builder.String()
}

var _ tea.Model = (*model)(nil)
