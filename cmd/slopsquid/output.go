package main

import (
	"fmt"
	"strings"

	"github.com/QRY91/slopsquid/internal/detector"
)

type crawlResult struct {
	URL    string               `json:"url"`
	Result *detector.ScanResult `json:"result"`
}

func printScanResult(path string, result *detector.ScanResult) {
	icon := ratingIcon(result.Rating)
	fmt.Printf("\n%s %s — score: %.0f/100 (%s)\n", icon, path, result.Score, result.Rating)

	// Group hits by type for cleaner output
	var words, trigrams, patterns []detector.Hit
	for _, h := range result.Hits {
		switch h.Type {
		case "word":
			words = append(words, h)
		case "trigram":
			trigrams = append(trigrams, h)
		case "pattern":
			patterns = append(patterns, h)
		}
	}

	if len(words) > 0 {
		printHitGroup("words", words)
	}
	if len(trigrams) > 0 {
		printHitGroup("phrases", trigrams)
	}
	if len(patterns) > 0 {
		printHitGroup("patterns", patterns)
	}

	fmt.Printf("  %d hits in %d lines, %d words — density: %.1f per 1k words\n",
		len(result.Hits), result.LineCount, result.WordCount, result.Density)
}

func printHitGroup(label string, hits []detector.Hit) {
	// Deduplicate by match text, count occurrences
	type matchInfo struct {
		hit   detector.Hit
		count int
		lines []int
	}
	seen := make(map[string]*matchInfo)
	var order []string

	for _, h := range hits {
		key := strings.ToLower(h.Match)
		if info, ok := seen[key]; ok {
			info.count++
			info.lines = append(info.lines, h.Line)
		} else {
			seen[key] = &matchInfo{hit: h, count: 1, lines: []int{h.Line}}
			order = append(order, key)
		}
	}

	for _, key := range order {
		info := seen[key]
		h := info.hit

		sevTag := severityTag(h.Severity)

		if info.count > 1 {
			fmt.Printf("  %s line %d (+%d): %q — %s\n",
				sevTag, info.lines[0], info.count-1, h.Match, h.Detail)
		} else {
			fmt.Printf("  %s line %d: %q — %s\n",
				sevTag, h.Line, h.Match, h.Detail)
		}
	}
}

func severityTag(severity string) string {
	switch severity {
	case "high":
		return "[!!]"
	case "medium":
		return "[! ]"
	case "low":
		return "[  ]"
	default:
		return "[  ]"
	}
}

func ratingIcon(rating string) string {
	switch rating {
	case "clean":
		return "."
	case "moderate":
		return "*"
	case "heavy":
		return "!"
	default:
		return "?"
	}
}

func printCrawlReport(source string, results []crawlResult, totalPages, skipped, totalWords, totalHits int) {
	fmt.Printf("\n== SlopSquid Report ==\n")
	fmt.Printf("   Source: %s\n", source)
	fmt.Printf("   Files scanned: %d (%d skipped)\n", totalPages, skipped)
	fmt.Printf("   Files scored: %d\n", len(results))
	fmt.Printf("   Total words: %d\n", totalWords)
	fmt.Printf("   Total hits: %d\n\n", totalHits)

	// Aggregate stats
	var clean, moderate, heavy int
	var totalScore float64
	for _, r := range results {
		totalScore += r.Result.Score
		switch r.Result.Rating {
		case "clean":
			clean++
		case "moderate":
			moderate++
		case "heavy":
			heavy++
		}
	}

	avgScore := 0.0
	if len(results) > 0 {
		avgScore = totalScore / float64(len(results))
	}

	fmt.Printf("   Breakdown: %d clean, %d moderate, %d heavy\n", clean, moderate, heavy)
	fmt.Printf("   Average score: %.1f/100\n\n", avgScore)

	// Top offenders (pages with hits, sorted by score desc — already sorted)
	fmt.Printf("-- Per-page scores --\n")
	for _, r := range results {
		icon := ratingIcon(r.Result.Rating)
		// Shorten URL for display
		path := r.URL
		if idx := strings.Index(path, "://"); idx != -1 {
			after := path[idx+3:]
			if slashIdx := strings.Index(after, "/"); slashIdx != -1 {
				path = after[slashIdx:]
			}
		}
		fmt.Printf("%s %5.1f  %-8s  %3d hits  %5d words  %s\n",
			icon, r.Result.Score, r.Result.Rating, len(r.Result.Hits), r.Result.WordCount, path)
	}

	// Top hits across entire site
	type siteHit struct {
		match    string
		hitType  string
		severity string
		detail   string
		count    int
		pages    []string
	}
	hitMap := make(map[string]*siteHit)
	var hitOrder []string

	for _, r := range results {
		for _, h := range r.Result.Hits {
			key := strings.ToLower(h.Match)
			if sh, ok := hitMap[key]; ok {
				sh.count++
				// Track unique pages
				found := false
				for _, p := range sh.pages {
					if p == r.URL {
						found = true
						break
					}
				}
				if !found {
					sh.pages = append(sh.pages, r.URL)
				}
			} else {
				hitMap[key] = &siteHit{
					match:    h.Match,
					hitType:  h.Type,
					severity: h.Severity,
					detail:   h.Detail,
					count:    1,
					pages:    []string{r.URL},
				}
				hitOrder = append(hitOrder, key)
			}
		}
	}

	if len(hitOrder) > 0 {
		// Sort by count descending
		type kv struct {
			key string
			sh  *siteHit
		}
		var sorted []kv
		for _, k := range hitOrder {
			sorted = append(sorted, kv{k, hitMap[k]})
		}
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].sh.count > sorted[i].sh.count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		fmt.Printf("\n-- Most frequent hits across site --\n")
		limit := 15
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for _, s := range sorted[:limit] {
			fmt.Printf("  %s %dx across %d pages: %q — %s\n",
				severityTag(s.sh.severity), s.sh.count, len(s.sh.pages), s.sh.match, s.sh.detail)
		}
	}

	fmt.Println()
}
