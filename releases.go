package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	// errMsg is a message that carries an error.
	// It is used to communicate errors between commands and the update function.
	// It can be thrown by any function of the `releases` file.
	errMsg error
	// gitReleaseExistsMsg is a message that carries information about
	// whether a GitHub release exists or not.
	gitReleaseExistsMsg struct {
		exists  bool
		release string
	}
	// gitReleasesDownloadSuccessMsg is a message that carries a list of GitHub releases.
	gitReleasesDownloadSuccessMsg = []Release
	// gitReleaseDownloadedMsg is a message that carries information about
	// a downloaded GitHub release: the release name, the destination directory,
	// and whether the result was cached or not.
	gitReleaseDownloadedMsg struct {
		release string
		dest    string
		cached  bool
	}
	// analysisDoneMsg is a message that carries information about the analysis
	// of a release. See AnalysisResult for more information.
	analysisDoneMsg = AnalysisResult
)

// AnalysisResult carries information about the analysis
// of a release: the total number of lines, the total number of files, and
// the number of lines by language, in addition to the release tag.
type AnalysisResult struct {
	releaseTag      string
	totalLines      uint
	totalFiles      uint
	linesByLanguage map[string]uint
}

type ListItem struct {
	previous *ListItem
	next     *ListItem
	AnalysisResult
}

func (l ListItem) Title() string {
	textForDiff := func(diff int) string {
		if diff > 0 {
			return successStyle.Render(fmt.Sprintf("+%d lines", diff))
		} else if diff < 0 {
			return errorStyle.Render(fmt.Sprintf("%d lines", diff))
		} else {
			return "No change"
		}
	}
	var sb strings.Builder

	if l.previous != nil {
		// All releases except the last one of the list
		sb.WriteString("  ")
		diffWithPrevious := int(l.totalLines) - int(l.previous.totalLines)
		sb.WriteString(textForDiff(diffWithPrevious))

		if l.next == nil {
			// First release of the list
			sb.WriteString(" • Total: ")
			first := l.previous
			for first.previous != nil {
				first = first.previous
			}
			diffWithFirst := int(l.totalLines) - int(first.totalLines)
			sb.WriteString(textForDiff(diffWithFirst))
		}
	}
	return l.releaseTag + sb.String()
}

func (l ListItem) Description() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d files • %d lines • ", l.totalFiles, l.totalLines))

	// Sort and shorten map
	type kv struct {
		Key   string
		Value uint
	}
	sorted := make([]kv, 0, len(l.linesByLanguage))
	for k, v := range l.linesByLanguage {
		sorted = append(sorted, kv{k, v})
	}
	slices.SortStableFunc(
		sorted, func(a, b kv) int {
			return cmp.Compare(a.Value, b.Value)
		},
	)
	visibleLanguages := 2
	if len(sorted) > visibleLanguages {
		// Shorten to visibleLanguages languages and concat all the others into the "Other" category
		otherElem := kv{fmt.Sprintf("and %d more", len(sorted[visibleLanguages:])), 0}
		for i := visibleLanguages; i < len(sorted); i++ {
			otherElem.Value += l.linesByLanguage[sorted[i].Key]
		}
		sorted = append(sorted[:visibleLanguages], otherElem)
	}

	// Print languages
	for i, lang := range sorted {
		if i > 0 {
			sb.WriteString(" / ")
		}
		sb.WriteString(fmt.Sprintf("%s (%d lines)", lang.Key, lang.Value))
	}

	return sb.String()
}

func (l ListItem) FilterValue() string {
	return l.releaseTag
}

var _ list.DefaultItem = (*ListItem)(nil)

// extToLang is a map that maps file extensions to programming languages.
// It is used to count the number of lines by language.
// It is not exhaustive and can be extended as needed.
// Note that keys should be lowercase, don't contain two-dot extensions,
// and start by a leading dot, in order to directly be used with filepath.Ext.
var extToLang = map[string]string{
	".js":   "JavaScript",
	".cjs":  "JavaScript",
	".mjs":  "JavaScript",
	".ts":   "TypeScript",
	".map":  "Source Map",
	".json": "JSON",
	".md":   "Markdown",
}

// DoesGitHubReleaseExist checks if a GitHub release exists for
// a given repository. Can use a token for authentication.
func DoesGitHubReleaseExist(ownerRepo, token, release string) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf(
				"https://api.github.com/repos/%s/releases/tags/%s",
				strings.TrimSuffix(ownerRepo, ".git"),
				release,
			),
			nil,
		)
		if err != nil {
			return errMsg(err)
		}

		req.Header.Add("Accept", "application/vnd.github+json")
		if token != "" {
			req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp.Body)

		if resp.StatusCode == http.StatusForbidden {
			return errMsg(fmt.Errorf("forbidden, please check your token or provide one"))
		}

		return gitReleaseExistsMsg{
			exists:  resp.StatusCode == http.StatusOK,
			release: release,
		}
	}
}

// GetGitHubReleases fetches GitHub releases for a repository.
// It can use a token for authentication, and it will fetch only
// releases between the `from` and the `to` release, ignoring the
// releases that don't match the `regex` regular expression.
func GetGitHubReleases(ownerRepo, token, from, to, regex string) tea.Cmd {
	page := 1
	fetchReleases := func() ([]Release, error) {
		request, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf(
				"https://api.github.com/repos/%s/releases",
				strings.TrimSuffix(ownerRepo, ".git"),
			),
			nil,
		)
		if err != nil {
			return nil, err
		}

		query := request.URL.Query()
		query.Add("page", fmt.Sprintf("%d", page))
		request.URL.RawQuery = query.Encode()

		request.Header.Add("Accept", "application/vnd.github+json")
		if token != "" {
			request.Header.Add("Authorization", fmt.Sprintf("token %s", token))
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				panic(err)
			}
		}(response.Body)

		if response.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("forbidden, please check your token or provide one")
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		var releases []Release
		err = json.Unmarshal(body, &releases)
		if err != nil {
			return releases, err
		}

		// Sort releases by reverse creation date
		slices.SortStableFunc(
			releases, func(a, b Release) int {
				return cmp.Compare(a.CreatedAt.Unix(), b.CreatedAt.Unix())
			},
		)

		return releases, nil
	}

	var compile *regexp.Regexp
	if regex != "" {
		var err error
		compile, err = regexp.Compile(regex)
		if err != nil {
			return func() tea.Msg {
				return errMsg(err)
			}
		}
	}

	return func() tea.Msg {
		var releases []Release

		foundFrom := false
		foundTo := false

		for {
			fetchedReleases, err := fetchReleases()
			if err != nil {
				return errMsg(err)
			}

			if releases == nil {
				// Slightly optimize the slice allocation
				releases = make([]Release, 0, len(fetchedReleases))
			}

			for _, release := range fetchedReleases {
				if compile != nil {
					if compile.MatchString(release.TagName) {
						continue
					}
				}
				if foundFrom && foundTo {
					// We've found both releases, so we don't need to add any anymore
					break
				}
				if release.TagName == from {
					foundFrom = true
				} else if release.TagName == to {
					foundTo = true
				}
				if !foundFrom && !foundTo {
					// Don't start adding releases until we find the first one
					continue
				}
				releases = append(releases, release)
			}

			if foundFrom && foundTo {
				// We've found both releases, so we don't need to fetch any anymore
				break
			}

			page++
		}

		return releases
	}
}

// DownloadGitHubRelease downloads a GitHub release from npmjs.com
// and extracts it to a destination directory.
// The destination directory is determined by the `destDir` function,
// which receives the release name as an argument.
func DownloadGitHubRelease(release, destDir string) tea.Cmd {
	return func() tea.Msg {
		// Create the destination directory
		dest := filepath.Clean(filepath.Join(destDir, release))
		if _, err := os.Stat(dest); err == nil {
			return gitReleaseDownloadedMsg{
				release: release,
				dest:    dest,
				cached:  true,
			}
		} else if err = os.MkdirAll(dest, 0750); err != nil {
			return errMsg(err)
		}

		// Create the URL
		// sveltejs/svelte svelte@5.0.0-next.90 -> https://registry.npmjs.com/svelte/-/svelte-5.0.0-next.90.tgz
		// sveltejs/kit @sveltejs/kit@1.0.0-next.589 -> https://registry.npmjs.com/@sveltejs/kit/-/kit-1.0.0-next.589.tgz
		name := ""
		if split := strings.Split(release, "@"); len(split) > 0 {
			if len(split) > 1 && strings.HasPrefix(release, "@") {
				name = "@" + split[1]
			} else {
				name = strings.Split(release, "@")[0]
			}
		}
		pkg := release
		if strings.Contains(release, "/") {
			pkg = strings.SplitN(release, "/", 2)[1]
		}
		url := fmt.Sprintf(
			"https://registry.npmjs.com/%s/-/%s.tgz",
			name, strings.ReplaceAll(pkg, "@", "-"),
		)

		// Fetch the release
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return errMsg(err)
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return errMsg(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(response.Body)

		if response.StatusCode != http.StatusOK {
			if response.StatusCode == http.StatusNotFound {
				return errMsg(fmt.Errorf("release not found at %s", request.URL.String()))
			}
			return errMsg(fmt.Errorf("could not download release: %s", response.Status))
		}

		// Un-tar the release
		err = Untar(dest, response.Body)
		if err != nil {
			return errMsg(err)
		}

		return gitReleaseDownloadedMsg{
			release: release,
			dest:    dest,
		}
	}
}

// AnalyzeRelease analyzes a release by counting lines of code
// for a given release within the location directory.
func AnalyzeRelease(locationDir string, releaseTag string) tea.Cmd {
	return func() tea.Msg {
		totalLines := uint(0)
		totalFiles := uint(0)
		linesByLanguage := make(map[string]uint)

		// Walk the directory
		err := filepath.WalkDir(
			filepath.Clean(filepath.Join(locationDir, releaseTag)),
			func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				// Count lines of code
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer func(file *os.File) {
					err = file.Close()
					if err != nil {
						panic(err)
					}
				}(file)

				lines, err := CountLines(file)
				if err != nil {
					return err
				}
				totalLines += lines
				totalFiles++

				// Count languages
				extension := filepath.Ext(path)
				if extension == "" {
					return nil
				}
				language := "Other"
				if lang, ok := extToLang[extension]; ok {
					language = lang
				}
				linesByLanguage[language] += lines

				return nil
			},
		)
		if err != nil {
			return errMsg(err)
		}

		return analysisDoneMsg{releaseTag, totalLines, totalFiles, linesByLanguage}
	}
}
