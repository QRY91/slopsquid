package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/QRY91/slopsquid/internal/detector"
	"github.com/QRY91/slopsquid/internal/scanner"
)

// ScanResult represents the result of scanning a file
type ScanResult struct {
	File   *scanner.FileInfo         `json:"file"`
	Result *detector.DetectionResult `json:"detection"`
}

// printTextResult prints a single scan result in human-readable format
func printTextResult(file *scanner.FileInfo, detection *detector.DetectionResult) {
	fmt.Printf("\n📄 %s (confidence: %.2f)\n", file.Path, detection.Confidence)

	if len(detection.Patterns) > 0 {
		fmt.Println("  🎯 Patterns detected:")
		for i, pattern := range detection.Patterns {
			fmt.Printf("    %d. %s: %s", i+1, pattern.Type, pattern.Pattern)
			if pattern.Suggestion != "" {
				fmt.Printf(" → %s", pattern.Suggestion)
			}
			fmt.Println()
			if pattern.Context != "" && len(pattern.Context) < 100 {
				fmt.Printf("       Context: %s\n", strings.TrimSpace(pattern.Context))
			}
		}
	}

	if len(detection.Suggestions) > 0 {
		fmt.Println("  💡 Suggestions:")
		for _, suggestion := range detection.Suggestions {
			fmt.Printf("     • %s\n", suggestion)
		}
	}
}

// outputJSON outputs results in JSON format
func outputJSON(results []ScanResult, outputFile string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write JSON file: %w", err)
		}
		fmt.Printf("🦑 Results written to %s\n", outputFile)
	} else {
		fmt.Println(string(data))
	}

	return nil
}

// outputYAML outputs results in YAML format
func outputYAML(results []ScanResult, outputFile string) error {
	// Simple YAML output - in a real implementation you'd use gopkg.in/yaml.v3
	var output strings.Builder

	for i, result := range results {
		if i > 0 {
			output.WriteString("---\n")
		}

		output.WriteString(fmt.Sprintf("file:\n"))
		output.WriteString(fmt.Sprintf("  path: %s\n", result.File.Path))
		output.WriteString(fmt.Sprintf("  size: %d\n", result.File.Size))
		output.WriteString(fmt.Sprintf("  type: %s\n", result.File.Type))

		output.WriteString(fmt.Sprintf("detection:\n"))
		output.WriteString(fmt.Sprintf("  confidence: %.3f\n", result.Result.Confidence))
		output.WriteString(fmt.Sprintf("  text_length: %d\n", result.Result.TextLength))

		if len(result.Result.Patterns) > 0 {
			output.WriteString("  patterns:\n")
			for _, pattern := range result.Result.Patterns {
				output.WriteString(fmt.Sprintf("    - type: %s\n", pattern.Type))
				output.WriteString(fmt.Sprintf("      pattern: %s\n", pattern.Pattern))
				output.WriteString(fmt.Sprintf("      confidence: %.3f\n", pattern.Confidence))
				if pattern.Context != "" {
					output.WriteString(fmt.Sprintf("      context: %s\n", pattern.Context))
				}
				if pattern.Suggestion != "" {
					output.WriteString(fmt.Sprintf("      suggestion: %s\n", pattern.Suggestion))
				}
			}
		}

		if len(result.Result.Suggestions) > 0 {
			output.WriteString("  suggestions:\n")
			for _, suggestion := range result.Result.Suggestions {
				output.WriteString(fmt.Sprintf("    - %s\n", suggestion))
			}
		}

		output.WriteString("\n")
	}

	yamlData := output.String()

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(yamlData), 0644)
		if err != nil {
			return fmt.Errorf("failed to write YAML file: %w", err)
		}
		fmt.Printf("🦑 Results written to %s\n", outputFile)
	} else {
		fmt.Print(yamlData)
	}

	return nil
}
