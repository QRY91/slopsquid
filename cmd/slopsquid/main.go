package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/QRY91/slopsquid/internal/crawler"
	"github.com/QRY91/slopsquid/internal/detector"
	"github.com/QRY91/slopsquid/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	version = "0.3.0"

	// Global flags
	verbose   bool
	recursive bool
	jsonOut   bool

	// Report flags
	reportDepth   int
	reportMax     int
	reportDelay   int
	reportWorkers int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "slopsquid",
	Short:   "Detect LLM-overused patterns in text",
	Version: version,
	Long: `SlopSquid scans text files for words, phrases, and structural patterns
that are statistically overrepresented in LLM output relative to human writing.

Based on frequency-ratio data from the Antislop paper (Paech et al., 2025),
which analyzed 67 AI models against human writing baselines.`,
}

var scanCmd = &cobra.Command{
	Use:   "scan [file|directory...]",
	Short: "Scan files and report slop hits with scoring",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(args)
	},
}

var scoreCmd = &cobra.Command{
	Use:   "score [file...]",
	Short: "Output slop density score (0-100) for each file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScore(args)
	},
}

var reportCmd = &cobra.Command{
	Use:   "report <url|directory>",
	Short: "Generate a consolidated slop report for a site or directory",
	Long: `Report generates a consolidated slop analysis. Accepts either:

  A URL:       slopsquid report qry.zone
               Crawls the site, respects robots.txt, seeds from sitemap.xml

  A directory: slopsquid report ./src/
               Scans local files recursively with HTML text extraction

Both produce the same report format: per-page scores, aggregate stats,
and the most frequent slop patterns across the entire corpus.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReport(args[0])
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&recursive, "recursive", "r", true, "process directories recursively")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output as JSON")

	reportCmd.Flags().IntVar(&reportDepth, "depth", 10, "maximum crawl depth (URLs only)")
	reportCmd.Flags().IntVar(&reportMax, "max-pages", 200, "maximum pages to crawl (URLs only)")
	reportCmd.Flags().IntVar(&reportDelay, "delay", 200, "delay between requests in ms (URLs only)")
	reportCmd.Flags().IntVar(&reportWorkers, "workers", 3, "concurrent requests (URLs only)")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(scoreCmd)
	rootCmd.AddCommand(reportCmd)
}

func getDetectorAndFiles(targets []string) (*detector.Detector, []*scanner.FileInfo, error) {
	d, err := detector.NewDetector()
	if err != nil {
		return nil, nil, fmt.Errorf("initializing detector: %w", err)
	}

	scanOptions := scanner.ScanOptions{
		Recursive:   recursive,
		MaxFileSize: 10 * 1024 * 1024,
	}
	fileScanner := scanner.NewScanner(scanOptions)

	files, err := fileScanner.ScanTargets(targets)
	if err != nil {
		return nil, nil, fmt.Errorf("scanning targets: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "found %d files to analyze\n", len(files))
	}

	return d, files, nil
}

func runScan(targets []string) error {
	d, files, err := getDetectorAndFiles(targets)
	if err != nil {
		return err
	}

	type fileResult struct {
		Path   string               `json:"path"`
		Result *detector.ScanResult `json:"result"`
	}

	var results []fileResult
	totalHits := 0

	for _, file := range files {
		if file.Error != "" {
			if verbose {
				fmt.Fprintf(os.Stderr, "skip %s: %s\n", file.Path, file.Error)
			}
			continue
		}
		if len(file.Content) < 20 {
			continue
		}

		result := d.Scan(file.Content)
		result.Path = file.Path

		if len(result.Hits) > 0 {
			results = append(results, fileResult{Path: file.Path, Result: result})
			totalHits += len(result.Hits)
		}
	}

	if jsonOut {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(results) == 0 {
		fmt.Printf("scanned %d files — clean\n", len(files))
		return nil
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Result.Score > results[j].Result.Score
	})

	for _, r := range results {
		printScanResult(r.Path, r.Result)
	}

	fmt.Printf("\n%d files scanned, %d with hits, %d total detections\n",
		len(files), len(results), totalHits)

	return nil
}

func runScore(targets []string) error {
	d, files, err := getDetectorAndFiles(targets)
	if err != nil {
		return err
	}

	type scoreEntry struct {
		Path    string  `json:"path"`
		Score   float64 `json:"score"`
		Rating  string  `json:"rating"`
		Hits    int     `json:"hits"`
		Words   int     `json:"words"`
		Density float64 `json:"density"`
	}

	var entries []scoreEntry

	for _, file := range files {
		if file.Error != "" || len(file.Content) < 20 {
			continue
		}

		result := d.Scan(file.Content)
		entries = append(entries, scoreEntry{
			Path:    file.Path,
			Score:   result.Score,
			Rating:  result.Rating,
			Hits:    len(result.Hits),
			Words:   result.WordCount,
			Density: result.Density,
		})
	}

	// Sort by score descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	if jsonOut {
		data, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	for _, e := range entries {
		icon := ratingIcon(e.Rating)
		fmt.Printf("%s %5.1f  %-8s  %3d hits  %5d words  %s\n",
			icon, e.Score, e.Rating, e.Hits, e.Words, e.Path)
	}

	return nil
}

// isURL returns true if the target looks like a URL rather than a local path
func isURL(target string) bool {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return true
	}
	// Bare domain: contains a dot but no path separator at start
	if strings.Contains(target, ".") && !strings.HasPrefix(target, ".") && !strings.HasPrefix(target, "/") {
		// Check it's not a relative file path like "foo.html"
		if _, err := os.Stat(target); os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func runReport(target string) error {
	if isURL(target) {
		return runReportHTTP(target)
	}
	return runReportLocal(target)
}

func runReportHTTP(rootURL string) error {
	if !strings.HasPrefix(rootURL, "http") {
		rootURL = "https://" + rootURL
	}

	d, err := detector.NewDetector()
	if err != nil {
		return fmt.Errorf("initializing detector: %w", err)
	}

	c, err := crawler.New(rootURL, crawler.Options{
		MaxDepth:    reportDepth,
		MaxPages:    reportMax,
		Concurrency: reportWorkers,
		Delay:       time.Duration(reportDelay) * time.Millisecond,
		Verbose:     verbose,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "crawling %s ...\n", rootURL)

	pages, err := c.Crawl(func(n int, url string) {
		if verbose {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", n, url)
		} else {
			fmt.Fprintf(os.Stderr, "\r  %d pages fetched", n)
		}
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\r  %d pages fetched\n", len(pages))

	var results []crawlResult
	var skipped int
	totalWords := 0
	totalHits := 0

	for _, page := range pages {
		if page.Error != "" || len(strings.Fields(page.Text)) < 10 {
			skipped++
			continue
		}

		result := d.Scan(page.Text)
		result.Path = page.URL
		totalWords += result.WordCount
		totalHits += len(result.Hits)

		results = append(results, crawlResult{URL: page.URL, Result: result})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Result.Score > results[j].Result.Score
	})

	if jsonOut {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	printCrawlReport(rootURL, results, len(pages), skipped, totalWords, totalHits)
	return nil
}

func runReportLocal(target string) error {
	absTarget, _ := filepath.Abs(target)

	d, err := detector.NewDetector()
	if err != nil {
		return fmt.Errorf("initializing detector: %w", err)
	}

	scanOptions := scanner.ScanOptions{
		Recursive:   true,
		MaxFileSize: 10 * 1024 * 1024,
	}
	fileScanner := scanner.NewScanner(scanOptions)

	files, err := fileScanner.ScanTargets([]string{target})
	if err != nil {
		return fmt.Errorf("scanning targets: %w", err)
	}

	fmt.Fprintf(os.Stderr, "scanning %s ...\n", absTarget)

	var results []crawlResult
	var skipped int
	totalWords := 0
	totalHits := 0

	for _, file := range files {
		if file.Error != "" {
			skipped++
			continue
		}

		text := file.Content
		// For HTML files, use the crawler's text extractor for consistency
		ext := strings.ToLower(filepath.Ext(file.Path))
		if ext == ".html" || ext == ".htm" {
			raw, readErr := os.ReadFile(file.Path)
			if readErr == nil {
				text = crawler.ExtractText(string(raw))
			}
		}

		if len(strings.Fields(text)) < 10 {
			skipped++
			continue
		}

		result := d.Scan(text)
		result.Path = file.Path
		totalWords += result.WordCount
		totalHits += len(result.Hits)

		results = append(results, crawlResult{URL: file.Path, Result: result})
	}

	fmt.Fprintf(os.Stderr, "  %d files processed\n", len(files))

	sort.Slice(results, func(i, j int) bool {
		return results[i].Result.Score > results[j].Result.Score
	})

	if jsonOut {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	printCrawlReport(absTarget, results, len(files), skipped, totalWords, totalHits)
	return nil
}
