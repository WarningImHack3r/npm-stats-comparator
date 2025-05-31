package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

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
	sb.WriteString(
		fmt.Sprintf(
			"%d files • %d lines • %s ",
			l.totalFiles, l.totalLines, ByteCountSI(l.totalDirSize),
		),
	)
	if l.tarSize > 0 {
		sb.WriteString(fmt.Sprintf("(%s gz) • ", ByteCountSI(l.tarSize)))
	} else {
		sb.WriteString("• ")
	}

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
			return cmp.Compare(b.Value, a.Value)
		},
	)
	visibleLanguages := 2
	if len(sorted) > visibleLanguages {
		// Shorten to visibleLanguages languages and concat all the others into the "Other" category
		otherElem := kv{fmt.Sprintf("%d other languages", len(sorted[visibleLanguages:])), 0}
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
