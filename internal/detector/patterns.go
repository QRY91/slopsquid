package detector

import (
	"fmt"
	"regexp"
	"strings"
)

// DetectionResult represents the result of AI pattern detection
type DetectionResult struct {
	Confidence  float64        `json:"confidence"`
	Patterns    []PatternMatch `json:"patterns"`
	TextLength  int            `json:"text_length"`
	Suggestions []string       `json:"suggestions,omitempty"`
}

// PatternMatch represents a specific pattern found in text
type PatternMatch struct {
	Type       string  `json:"type"`
	Pattern    string  `json:"pattern"`
	Position   int     `json:"position"`
	Length     int     `json:"length"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context"`
	Suggestion string  `json:"suggestion,omitempty"`
}

// PatternDetector is the main AI pattern detection engine
type PatternDetector struct {
	sensitivity float64
	patterns    *PatternDatabase
}

// PatternDatabase contains all the patterns used for detection
type PatternDatabase struct {
	aiPhrases          []string
	corporateSpeak     []string
	buzzwords          []string
	formalConnectors   []string
	aiSpecificPatterns []*regexp.Regexp
	syntaxPatterns     []*regexp.Regexp
}

// NewPatternDetector creates a new detector with specified sensitivity
func NewPatternDetector(sensitivity float64) *PatternDetector {
	return &PatternDetector{
		sensitivity: sensitivity,
		patterns:    buildPatternDatabase(),
	}
}

// DetectAIPatterns analyzes text and returns detection results
func (pd *PatternDetector) DetectAIPatterns(text string) *DetectionResult {
	result := &DetectionResult{
		TextLength: len(text),
		Patterns:   make([]PatternMatch, 0),
	}

	// Skip very short text
	if len(text) < 50 {
		return result
	}

	lowerText := strings.ToLower(text)
	sentences := splitSentences(text)

	// Detect various pattern types
	score := 0.0

	// 1. AI-specific phrases
	score += pd.detectAIPhrases(lowerText, result)

	// 2. Corporate speak and buzzwords
	score += pd.detectCorporateSpeak(lowerText, result)

	// 3. Formal connectors overuse
	score += pd.detectFormalConnectors(lowerText, result)

	// 4. Sentence structure patterns
	score += pd.detectSentencePatterns(sentences, result)

	// 5. AI-specific regex patterns
	score += pd.detectRegexPatterns(text, lowerText, result)

	// 6. Repetitive language patterns
	score += pd.detectRepetitivePatterns(text, result)

	// 7. Unnatural formality
	score += pd.detectUnnaturalFormality(lowerText, sentences, result)

	result.Confidence = min(score, 1.0)

	// Add general suggestions if confidence is high
	if result.Confidence > 0.7 {
		result.Suggestions = pd.generateSuggestions(result)
	}

	return result
}

func (pd *PatternDetector) detectAIPhrases(lowerText string, result *DetectionResult) float64 {
	score := 0.0
	for _, phrase := range pd.patterns.aiPhrases {
		if index := strings.Index(lowerText, phrase); index != -1 {
			match := PatternMatch{
				Type:       "ai_phrase",
				Pattern:    phrase,
				Position:   index,
				Length:     len(phrase),
				Confidence: 0.8,
				Context:    extractContext(lowerText, index, len(phrase)),
				Suggestion: getPhraseSuggestion(phrase),
			}
			result.Patterns = append(result.Patterns, match)
			score += 0.15
		}
	}
	return score
}

func (pd *PatternDetector) detectCorporateSpeak(lowerText string, result *DetectionResult) float64 {
	score := 0.0

	// Check for buzzwords
	for _, buzzword := range pd.patterns.buzzwords {
		count := strings.Count(lowerText, buzzword)
		if count > 0 {
			// Find first occurrence for position
			index := strings.Index(lowerText, buzzword)
			match := PatternMatch{
				Type:       "buzzword",
				Pattern:    buzzword,
				Position:   index,
				Length:     len(buzzword),
				Confidence: 0.6,
				Context:    extractContext(lowerText, index, len(buzzword)),
			}
			result.Patterns = append(result.Patterns, match)
			score += float64(count) * 0.1
		}
	}

	// Check for corporate speak
	for _, phrase := range pd.patterns.corporateSpeak {
		if index := strings.Index(lowerText, phrase); index != -1 {
			match := PatternMatch{
				Type:       "corporate_speak",
				Pattern:    phrase,
				Position:   index,
				Length:     len(phrase),
				Confidence: 0.7,
				Context:    extractContext(lowerText, index, len(phrase)),
			}
			result.Patterns = append(result.Patterns, match)
			score += 0.12
		}
	}

	return min(score, 0.4) // Cap corporate speak contribution
}

func (pd *PatternDetector) detectFormalConnectors(lowerText string, result *DetectionResult) float64 {
	connectorCount := 0
	for _, connector := range pd.patterns.formalConnectors {
		count := strings.Count(lowerText, connector)
		if count > 0 {
			connectorCount += count
			// Find first occurrence
			index := strings.Index(lowerText, connector)
			match := PatternMatch{
				Type:       "formal_connector",
				Pattern:    connector,
				Position:   index,
				Length:     len(connector),
				Confidence: 0.5,
				Context:    extractContext(lowerText, index, len(connector)),
			}
			result.Patterns = append(result.Patterns, match)
		}
	}

	// Score based on density of formal connectors
	wordCount := len(strings.Fields(lowerText))
	if wordCount > 0 {
		density := float64(connectorCount) / float64(wordCount)
		return min(density*2.0, 0.3) // Scale and cap
	}
	return 0.0
}

func (pd *PatternDetector) detectSentencePatterns(sentences []string, result *DetectionResult) float64 {
	if len(sentences) < 3 {
		return 0.0
	}

	score := 0.0
	perfectLengthCount := 0
	longSentenceCount := 0

	for _, sentence := range sentences {
		words := strings.Fields(strings.TrimSpace(sentence))
		wordCount := len(words)

		// AI tends to write consistently medium-length sentences
		if wordCount >= 15 && wordCount <= 25 {
			perfectLengthCount++
		}

		// AI also creates many long, complex sentences
		if wordCount > 25 {
			longSentenceCount++
		}
	}

	totalSentences := len(sentences)
	perfectRatio := float64(perfectLengthCount) / float64(totalSentences)
	longRatio := float64(longSentenceCount) / float64(totalSentences)

	if perfectRatio > 0.6 {
		result.Patterns = append(result.Patterns, PatternMatch{
			Type:       "sentence_uniformity",
			Pattern:    "uniform sentence lengths",
			Confidence: perfectRatio,
			Context:    "sentences are suspiciously uniform in length",
		})
		score += perfectRatio * 0.25
	}

	if longRatio > 0.4 {
		result.Patterns = append(result.Patterns, PatternMatch{
			Type:       "long_sentences",
			Pattern:    "excessive long sentences",
			Confidence: longRatio,
			Context:    "many sentences are unusually long and complex",
		})
		score += longRatio * 0.2
	}

	return score
}

func (pd *PatternDetector) detectRegexPatterns(text, lowerText string, result *DetectionResult) float64 {
	score := 0.0

	for _, pattern := range pd.patterns.aiSpecificPatterns {
		matches := pattern.FindAllStringIndex(lowerText, -1)
		for _, match := range matches {
			start, end := match[0], match[1]
			patternMatch := PatternMatch{
				Type:       "ai_pattern",
				Pattern:    pattern.String(),
				Position:   start,
				Length:     end - start,
				Confidence: 0.7,
				Context:    extractContext(lowerText, start, end-start),
			}
			result.Patterns = append(result.Patterns, patternMatch)
			score += 0.1
		}
	}

	return min(score, 0.3)
}

func (pd *PatternDetector) detectRepetitivePatterns(text string, result *DetectionResult) float64 {
	words := strings.Fields(strings.ToLower(text))
	if len(words) < 20 {
		return 0.0
	}

	wordCount := make(map[string]int)
	for _, word := range words {
		// Skip common words
		if len(word) > 4 && !isCommonWord(word) {
			wordCount[word]++
		}
	}

	repetitiveScore := 0.0
	for word, count := range wordCount {
		if count > 3 { // Word appears more than 3 times
			ratio := float64(count) / float64(len(words))
			if ratio > 0.02 { // More than 2% of text
				result.Patterns = append(result.Patterns, PatternMatch{
					Type:       "repetitive_word",
					Pattern:    word,
					Confidence: ratio * 5, // Scale up
					Context:    fmt.Sprintf("word '%s' repeated %d times", word, count),
				})
				repetitiveScore += ratio * 2
			}
		}
	}

	return min(repetitiveScore, 0.25)
}

func (pd *PatternDetector) detectUnnaturalFormality(lowerText string, sentences []string, result *DetectionResult) float64 {
	// Check for excessive use of passive voice
	passiveCount := 0
	passivePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b(is|are|was|were|being|been)\s+\w+ed\b`),
		regexp.MustCompile(`\b(is|are|was|were)\s+(being\s+)?\w+ed\s+by\b`),
	}

	for _, pattern := range passivePatterns {
		matches := pattern.FindAllString(lowerText, -1)
		passiveCount += len(matches)
	}

	if len(sentences) > 0 {
		passiveRatio := float64(passiveCount) / float64(len(sentences))
		if passiveRatio > 0.3 {
			result.Patterns = append(result.Patterns, PatternMatch{
				Type:       "excessive_passive",
				Pattern:    "passive voice overuse",
				Confidence: passiveRatio,
				Context:    "excessive use of passive voice",
			})
			return passiveRatio * 0.2
		}
	}

	return 0.0
}

func (pd *PatternDetector) generateSuggestions(result *DetectionResult) []string {
	suggestions := make([]string, 0)

	if result.Confidence > 0.9 {
		suggestions = append(suggestions, "Consider rewriting this content to use more natural, conversational language")
	}
	if result.Confidence > 0.8 {
		suggestions = append(suggestions, "Try to vary sentence length and structure")
		suggestions = append(suggestions, "Replace formal connectors with simpler transitions")
	}
	if result.Confidence > 0.7 {
		suggestions = append(suggestions, "Remove unnecessary buzzwords and corporate jargon")
	}

	return suggestions
}

// Helper functions
func buildPatternDatabase() *PatternDatabase {
	return &PatternDatabase{
		aiPhrases: []string{
			"as an ai", "i apologize", "i understand", "i'm sorry", "i cannot",
			"large language model", "trained on", "variety of tasks",
			"human-like responses", "wide range of", "diverse set of",
			"it's worth noting", "it is important to note", "delve into",
		},
		corporateSpeak: []string{
			"leverage", "synergy", "paradigm", "ecosystem", "holistic",
			"streamline", "optimize", "maximize", "facilitate", "utilize",
			"cutting-edge", "state-of-the-art", "innovative", "robust",
			"comprehensive", "seamless", "scalable", "strategic",
		},
		buzzwords: []string{
			"disruptive", "revolutionary", "game-changing", "next-generation",
			"best-in-class", "world-class", "industry-leading", "market-leading",
			"transformative", "groundbreaking", "pioneering", "advanced",
		},
		formalConnectors: []string{
			"furthermore", "moreover", "subsequently", "therefore", "however",
			"nevertheless", "nonetheless", "additionally", "consequently",
			"thus", "hence", "whereby", "wherein", "thereof", "thereby",
		},
		aiSpecificPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(can be used for|is able to|has been trained|is capable of)\b`),
			regexp.MustCompile(`\b(a variety of|wide range of|diverse set of|massive dataset)\b`),
			regexp.MustCompile(`\b(language model|text generation|natural language processing)\b`),
			regexp.MustCompile(`\b(it's important to (note|understand)|it should be noted)\b`),
		},
	}
}

func splitSentences(text string) []string {
	// Simple sentence splitting
	sentences := regexp.MustCompile(`[.!?]+`).Split(text, -1)
	result := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 10 { // Skip very short fragments
			result = append(result, trimmed)
		}
	}
	return result
}

func extractContext(text string, position, length int) string {
	start := maxInt(0, position-30)
	end := minInt(len(text), position+length+30)
	return strings.TrimSpace(text[start:end])
}

func getPhraseSuggestion(phrase string) string {
	suggestions := map[string]string{
		"it's worth noting":       "consider",
		"it is important to note": "note that",
		"delve into":              "explore",
		"facilitate":              "help",
		"utilize":                 "use",
		"leverage":                "use",
		"comprehensive":           "complete",
		"robust":                  "strong",
		"seamless":                "smooth",
	}
	return suggestions[phrase]
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "that": true, "have": true, "for": true,
		"not": true, "with": true, "you": true, "this": true, "but": true,
		"his": true, "from": true, "they": true, "she": true, "her": true,
		"been": true, "than": true, "its": true, "who": true, "oil": true,
	}
	return commonWords[word]
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
