package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/QRY91/slopsquid/internal/detector"
	"github.com/QRY91/slopsquid/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"

	// Global flags
	verbose    bool
	confidence float64
	threads    int
	recursive  bool
	dryRun     bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "slopsquid",
	Short: "CLI tool for detecting and cleaning AI artifacts in documentation 🦑",
	Long: `SlopSquid is a CLI tool that detects and cleans AI-generated patterns
from your documentation, keeping your content authentic and polished.

Deploy the tentacles to identify corporate speak, buzzwords, and artificial
writing patterns in your files.`,
	Version: version,
}

var scanCmd = &cobra.Command{
	Use:   "scan [file/directory]",
	Short: "Scan files for AI-generated content patterns",
	Long: `Scan analyzes files or directories to detect AI-generated content patterns.
It identifies corporate speak, buzzwords, repetitive structures, and other
artificial writing signatures.

Examples:
  slopsquid scan README.md
  slopsquid scan docs/ --recursive --threads=8
  slopsquid scan --confidence=0.8 project/`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(args)
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean [file/directory]",
	Short: "Clean AI artifacts from files",
	Long: `Clean removes or suggests replacements for detected AI patterns in your files.
It can run in interactive mode for review or automatically fix obvious patterns.

Examples:
  slopsquid clean README.md --interactive
  slopsquid clean docs/ --auto-fix --safe
  slopsquid clean project/ --dry-run --confidence=0.9`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClean(args)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Float64VarP(&confidence, "confidence", "c", 0.7, "minimum confidence threshold (0.0-1.0)")
	rootCmd.PersistentFlags().IntVarP(&threads, "threads", "t", 4, "number of processing threads")
	rootCmd.PersistentFlags().BoolVarP(&recursive, "recursive", "r", false, "process directories recursively")

	// Scan command flags
	scanCmd.Flags().StringP("format", "f", "text", "output format (text, json, yaml)")
	scanCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
	scanCmd.Flags().Bool("git-staged", false, "scan only git staged files")

	// Clean command flags
	cleanCmd.Flags().BoolP("interactive", "i", false, "interactive mode - review each detection")
	cleanCmd.Flags().Bool("auto-fix", false, "automatically fix obvious patterns")
	cleanCmd.Flags().Bool("safe", false, "only apply safe, low-risk fixes")
	cleanCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes without applying them")
	cleanCmd.Flags().String("backup-suffix", ".bak", "suffix for backup files")

	// Add commands to root
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(cleanCmd)
}

func runScan(targets []string) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "🦑 SlopSquid scanning with confidence threshold %.2f\n", confidence)
		fmt.Fprintf(os.Stderr, "📁 Targets: %v\n", targets)
		if recursive {
			fmt.Fprintf(os.Stderr, "🔄 Recursive mode enabled\n")
		}
		fmt.Fprintf(os.Stderr, "🧵 Using %d processing threads\n", threads)
	}

	// Create scanner with options
	scanOptions := scanner.ScanOptions{
		Recursive:   recursive,
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
	fileScanner := scanner.NewScanner(scanOptions)

	// Create detector
	patternDetector := detector.NewPatternDetector(confidence)

	// Scan files
	files, err := fileScanner.ScanTargets(targets)
	if err != nil {
		return fmt.Errorf("scanning failed: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "📄 Found %d files to analyze\n", len(files))
	}

	// Results storage

	var results []ScanResult
	totalDetections := 0

	// Process each file
	for _, file := range files {
		if file.Error != "" {
			if verbose {
				fmt.Fprintf(os.Stderr, "⚠️  Error with %s: %s\n", file.Path, file.Error)
			}
			continue
		}

		if len(file.Content) < 50 {
			continue // Skip very short files
		}

		// Detect AI patterns
		detection := patternDetector.DetectAIPatterns(file.Content)

		if detection.Confidence >= confidence {
			results = append(results, ScanResult{
				File:   file,
				Result: detection,
			})
			totalDetections++

			// Print immediate feedback for text output - will be handled by caller
		}
	}

	// Return results for caller to handle formatting
	for _, result := range results {
		printTextResult(result.File, result.Result)
	}

	if len(results) == 0 {
		fmt.Printf("🦑 Scanned %d files, no AI patterns found above %.2f confidence\n",
			len(files), confidence)
	}

	return nil
}

func runClean(targets []string) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "🦑 SlopSquid cleaning with confidence threshold %.2f\n", confidence)
		fmt.Fprintf(os.Stderr, "📁 Targets: %v\n", targets)
		if dryRun {
			fmt.Fprintf(os.Stderr, "👀 Dry run mode - no changes will be made\n")
		}
	}

	// First scan to find issues
	scanOptions := scanner.ScanOptions{
		Recursive:   recursive,
		MaxFileSize: 10 * 1024 * 1024,
	}
	fileScanner := scanner.NewScanner(scanOptions)
	patternDetector := detector.NewPatternDetector(confidence)

	files, err := fileScanner.ScanTargets(targets)
	if err != nil {
		return fmt.Errorf("scanning failed: %w", err)
	}

	// Default values for now - will be passed as parameters later
	interactive := false
	autoFix := false
	safe := false

	filesProcessed := 0
	patternsFound := 0

	for _, file := range files {
		if file.Error != "" {
			continue
		}

		detection := patternDetector.DetectAIPatterns(file.Content)
		if detection.Confidence < confidence {
			continue
		}

		filesProcessed++
		patternsFound += len(detection.Patterns)

		fmt.Printf("\n📄 %s (confidence: %.2f)\n", file.Path, detection.Confidence)

		// Show detected patterns
		for i, pattern := range detection.Patterns {
			fmt.Printf("  %d. %s: %s", i+1, pattern.Type, pattern.Pattern)
			if pattern.Suggestion != "" {
				fmt.Printf(" → %s", pattern.Suggestion)
			}
			fmt.Println()
			if pattern.Context != "" {
				fmt.Printf("     Context: %s\n", pattern.Context)
			}
		}

		// Show suggestions
		if len(detection.Suggestions) > 0 {
			fmt.Println("  💡 Suggestions:")
			for _, suggestion := range detection.Suggestions {
				fmt.Printf("     • %s\n", suggestion)
			}
		}

		if interactive && !dryRun {
			fmt.Print("  Apply fixes? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				fmt.Printf("  ✅ Would apply fixes to %s (not implemented yet)\n", file.Path)
			}
		} else if autoFix && !dryRun {
			if !safe || detection.Confidence > 0.9 {
				fmt.Printf("  🔧 Would auto-fix %s (not implemented yet)\n", file.Path)
			}
		}
	}

	fmt.Printf("\n🦑 Summary: processed %d files, found %d patterns\n", filesProcessed, patternsFound)
	if dryRun {
		fmt.Println("👀 Dry run complete - no files were modified")
	}

	return nil
}
