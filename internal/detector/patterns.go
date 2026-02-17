package detector

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed data
var dataFS embed.FS

// Banlist data structures
type WordEntry struct {
	Word      string  `json:"word"`
	PctModels float64 `json:"pct_models"`
	Severity  string  `json:"severity"`
	Note      string  `json:"note,omitempty"`
}

type WordData struct {
	Words []WordEntry `json:"words"`
}

type TrigramEntry struct {
	Phrase    string  `json:"phrase"`
	PctModels float64 `json:"pct_models"`
	Severity  string  `json:"severity"`
	Note      string  `json:"note,omitempty"`
}

type TrigramData struct {
	Trigrams []TrigramEntry `json:"trigrams"`
}

type PatternEntry struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	OveruseRat  float64 `json:"overuse_ratio"`
	Regex       string  `json:"regex"`
	Note        string  `json:"note,omitempty"`
}

type PatternData struct {
	Patterns []PatternEntry `json:"patterns"`
}

// Hit represents a single slop detection
type Hit struct {
	Line     int     `json:"line"`
	Column   int     `json:"column"`
	Match    string  `json:"match"`
	Type     string  `json:"type"` // "word", "trigram", "pattern"
	Detail   string  `json:"detail"`
	Severity string  `json:"severity"`
	Weight   float64 `json:"weight"` // frequency-ratio based weight
}

// ScanResult is the result of scanning a single file
type ScanResult struct {
	Path      string  `json:"path"`
	Hits      []Hit   `json:"hits"`
	Score     float64 `json:"score"`
	LineCount int     `json:"line_count"`
	WordCount int     `json:"word_count"`
	Density   float64 `json:"density"` // hits per 1000 words
	Rating    string  `json:"rating"`  // "clean", "moderate", "heavy"
}

// Detector is the main slop detection engine
type Detector struct {
	words    []WordEntry
	trigrams []TrigramEntry
	patterns []compiledPattern
}

type compiledPattern struct {
	entry PatternEntry
	regex *regexp.Regexp
}

// PresetData holds the combined words/trigrams/patterns from a preset file
type PresetData struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Words       []WordEntry    `json:"words"`
	Trigrams    []TrigramEntry `json:"trigrams"`
	Patterns    []PatternEntry `json:"patterns"`
}

// DetectorOptions configures preset loading
type DetectorOptions struct {
	Presets   []string // Preset names or file paths
	PresetDir string   // External directory to search for presets by name
}

// NewDetector creates a detector with embedded banlist data
func NewDetector() (*Detector, error) {
	return NewDetectorWithPresets(nil)
}

// NewDetectorWithPresets creates a detector with base data plus optional preset names
func NewDetectorWithPresets(presets []string) (*Detector, error) {
	return NewDetectorWithOptions(DetectorOptions{Presets: presets})
}

// NewDetectorWithOptions creates a detector with full configuration
func NewDetectorWithOptions(opts DetectorOptions) (*Detector, error) {
	d := &Detector{}

	// Load word data
	wordBytes, err := dataFS.ReadFile("data/words.json")
	if err != nil {
		return nil, fmt.Errorf("loading word data: %w", err)
	}
	var wordData WordData
	if err := json.Unmarshal(wordBytes, &wordData); err != nil {
		return nil, fmt.Errorf("parsing word data: %w", err)
	}
	d.words = wordData.Words

	// Load trigram data
	trigramBytes, err := dataFS.ReadFile("data/trigrams.json")
	if err != nil {
		return nil, fmt.Errorf("loading trigram data: %w", err)
	}
	var trigramData TrigramData
	if err := json.Unmarshal(trigramBytes, &trigramData); err != nil {
		return nil, fmt.Errorf("parsing trigram data: %w", err)
	}
	d.trigrams = trigramData.Trigrams

	// Load and compile base pattern data
	patternBytes, err := dataFS.ReadFile("data/patterns.json")
	if err != nil {
		return nil, fmt.Errorf("loading pattern data: %w", err)
	}
	var patternData PatternData
	if err := json.Unmarshal(patternBytes, &patternData); err != nil {
		return nil, fmt.Errorf("parsing pattern data: %w", err)
	}

	for _, p := range patternData.Patterns {
		re, err := regexp.Compile("(?i)" + p.Regex)
		if err != nil {
			fmt.Printf("warning: skipping invalid pattern %q: %v\n", p.Name, err)
			continue
		}
		d.patterns = append(d.patterns, compiledPattern{entry: p, regex: re})
	}

	// Load presets (additive)
	for _, name := range opts.Presets {
		if err := d.loadPreset(name, opts.PresetDir); err != nil {
			return nil, fmt.Errorf("loading preset %q: %w", name, err)
		}
	}

	return d, nil
}

// loadPreset resolves a preset name or path and adds its data to the detector.
// Resolution order:
//  1. If name contains / or ends in .json, treat as a file path
//  2. If presetDir is set, look for <presetDir>/<name>.json
//  3. Look for embedded preset data/presets/<name>.json
func (d *Detector) loadPreset(name string, presetDir string) error {
	presetBytes, err := resolvePreset(name, presetDir)
	if err != nil {
		return err
	}

	var preset PresetData
	if err := json.Unmarshal(presetBytes, &preset); err != nil {
		return fmt.Errorf("parsing preset %q: %w", name, err)
	}

	d.words = append(d.words, preset.Words...)
	d.trigrams = append(d.trigrams, preset.Trigrams...)

	for _, p := range preset.Patterns {
		re, err := regexp.Compile("(?i)" + p.Regex)
		if err != nil {
			fmt.Printf("warning: skipping invalid preset pattern %q: %v\n", p.Name, err)
			continue
		}
		d.patterns = append(d.patterns, compiledPattern{entry: p, regex: re})
	}

	return nil
}

func resolvePreset(name string, presetDir string) ([]byte, error) {
	// 1. Direct file path
	if strings.Contains(name, "/") || strings.HasSuffix(name, ".json") {
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("reading preset file %q: %w", name, err)
		}
		return data, nil
	}

	// 2. External preset directory
	if presetDir != "" {
		extPath := filepath.Join(presetDir, name+".json")
		if data, err := os.ReadFile(extPath); err == nil {
			return data, nil
		}
	}

	// 3. Embedded preset
	embeddedPath := "data/presets/" + name + ".json"
	data, err := dataFS.ReadFile(embeddedPath)
	if err != nil {
		if presetDir != "" {
			return nil, fmt.Errorf("preset %q not found (checked %s and embedded data)", name, filepath.Join(presetDir, name+".json"))
		}
		return nil, fmt.Errorf("preset %q not found in embedded data", name)
	}
	return data, nil
}

// ListPresets returns the names of all available embedded presets
func ListPresets() ([]string, error) {
	entries, err := dataFS.ReadDir("data/presets")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

// PresetDescription returns the description of an embedded preset
func PresetDescription(name string) (string, error) {
	path := "data/presets/" + name + ".json"
	data, err := dataFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	var p PresetData
	if err := json.Unmarshal(data, &p); err != nil {
		return "", err
	}
	return p.Description, nil
}

// Scan analyzes text and returns slop hits with scoring
func (d *Detector) Scan(text string) *ScanResult {
	result := &ScanResult{}

	lines := strings.Split(text, "\n")
	result.LineCount = len(lines)
	result.WordCount = len(strings.Fields(text))

	if result.WordCount == 0 {
		result.Rating = "clean"
		return result
	}

	lowerText := strings.ToLower(text)

	// Build a line index for mapping character positions to line numbers
	lineOffsets := buildLineOffsets(text)

	// 1. Scan for banlist words
	for _, w := range d.words {
		d.scanWord(lowerText, w, lineOffsets, result)
	}

	// 2. Scan for trigrams
	for _, t := range d.trigrams {
		d.scanTrigram(lowerText, t, lineOffsets, result)
	}

	// 3. Scan for structural patterns
	for _, p := range d.patterns {
		d.scanPattern(lowerText, p, lineOffsets, result)
	}

	// Calculate score
	result.Score = d.calculateScore(result)
	result.Density = float64(len(result.Hits)) / float64(result.WordCount) * 1000

	if result.Score < 20 {
		result.Rating = "clean"
	} else if result.Score < 50 {
		result.Rating = "moderate"
	} else {
		result.Rating = "heavy"
	}

	return result
}

func (d *Detector) scanWord(lowerText string, w WordEntry, lineOffsets []int, result *ScanResult) {
	// Match whole words only
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(w.Word) + `\b`)
	matches := pattern.FindAllStringIndex(lowerText, -1)

	for _, m := range matches {
		line, col := posToLineCol(m[0], lineOffsets)
		result.Hits = append(result.Hits, Hit{
			Line:     line,
			Column:   col,
			Match:    lowerText[m[0]:m[1]],
			Type:     "word",
			Detail:   fmt.Sprintf("%.1f%% of models overuse this word", w.PctModels),
			Severity: w.Severity,
			Weight:   w.PctModels / 100.0,
		})
	}
}

func (d *Detector) scanTrigram(lowerText string, t TrigramEntry, lineOffsets []int, result *ScanResult) {
	// Trigrams need fuzzy matching — words may not be adjacent
	// Search for all words within a 5-word window
	words := strings.Fields(t.Phrase)
	if len(words) < 2 {
		return
	}

	// Simple substring search for the phrase
	idx := 0
	for {
		pos := strings.Index(lowerText[idx:], words[0])
		if pos == -1 {
			break
		}
		absPos := idx + pos

		// Check if remaining words appear nearby (within 60 chars)
		window := 60
		end := absPos + window
		if end > len(lowerText) {
			end = len(lowerText)
		}
		snippet := lowerText[absPos:end]

		allFound := true
		for _, w := range words[1:] {
			if !strings.Contains(snippet, w) {
				allFound = false
				break
			}
		}

		if allFound {
			line, col := posToLineCol(absPos, lineOffsets)
			result.Hits = append(result.Hits, Hit{
				Line:     line,
				Column:   col,
				Match:    t.Phrase,
				Type:     "trigram",
				Detail:   fmt.Sprintf("%.1f%% of models overuse this phrase", t.PctModels),
				Severity: t.Severity,
				Weight:   t.PctModels / 100.0,
			})
		}

		idx = absPos + len(words[0])
		if idx >= len(lowerText) {
			break
		}
	}
}

func (d *Detector) scanPattern(lowerText string, p compiledPattern, lineOffsets []int, result *ScanResult) {
	matches := p.regex.FindAllStringIndex(lowerText, -1)

	for _, m := range matches {
		line, col := posToLineCol(m[0], lineOffsets)
		matchText := lowerText[m[0]:m[1]]
		if len(matchText) > 60 {
			matchText = matchText[:60] + "..."
		}
		result.Hits = append(result.Hits, Hit{
			Line:     line,
			Column:   col,
			Match:    matchText,
			Type:     "pattern",
			Detail:   fmt.Sprintf("%s (%.1fx overrepresented)", p.entry.Description, p.entry.OveruseRat),
			Severity: p.entry.Severity,
			Weight:   p.entry.OveruseRat / 10.0, // normalize to ~0-1 range
		})
	}
}

func (d *Detector) calculateScore(result *ScanResult) float64 {
	if len(result.Hits) == 0 || result.WordCount == 0 {
		return 0
	}

	totalWeight := 0.0
	for _, h := range result.Hits {
		totalWeight += h.Weight
	}

	// Score = weighted hits per 1000 words, scaled to 0-100
	rawScore := (totalWeight / float64(result.WordCount)) * 1000

	// Clamp to 0-100
	if rawScore > 100 {
		return 100
	}
	return rawScore
}

// Helper functions
func buildLineOffsets(text string) []int {
	offsets := []int{0}
	for i, ch := range text {
		if ch == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

func posToLineCol(pos int, lineOffsets []int) (int, int) {
	line := 1
	for i, offset := range lineOffsets {
		if offset > pos {
			break
		}
		line = i + 1
	}
	col := pos - lineOffsets[line-1] + 1
	return line, col
}
