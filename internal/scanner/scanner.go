package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileInfo represents information about a scanned file
type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ScanOptions configures the scanner behavior
type ScanOptions struct {
	Recursive    bool
	MaxFileSize  int64
	Extensions   []string
	ExcludeDirs  []string
	ExcludeFiles []string
	FollowLinks  bool
}

// Scanner handles file discovery and content extraction
type Scanner struct {
	options ScanOptions
}

// NewScanner creates a new file scanner with the given options
func NewScanner(options ScanOptions) *Scanner {
	if options.MaxFileSize == 0 {
		options.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	if len(options.Extensions) == 0 {
		options.Extensions = []string{
			".md", ".markdown", ".txt", ".text",
			".rst", ".adoc", ".asciidoc",
			".html", ".htm", ".xml",
		}
	}

	if len(options.ExcludeDirs) == 0 {
		options.ExcludeDirs = []string{
			".git", ".svn", ".hg", ".bzr",
			"node_modules", ".vscode", ".idea",
			"vendor", "target", "build", "dist",
		}
	}

	return &Scanner{options: options}
}

// ScanTargets processes a list of file or directory targets
func (s *Scanner) ScanTargets(targets []string) ([]*FileInfo, error) {
	var allFiles []*FileInfo
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			files, err := s.scanTarget(target)
			if err != nil {
				// Log error but continue with other targets
				fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", target, err)
				return
			}

			mu.Lock()
			allFiles = append(allFiles, files...)
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	return allFiles, nil
}

// scanTarget processes a single target (file or directory)
func (s *Scanner) scanTarget(target string) ([]*FileInfo, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", target, err)
	}

	if info.IsDir() {
		return s.scanDirectory(target)
	}

	file, err := s.scanFile(target, info)
	if err != nil {
		return nil, err
	}

	if file != nil {
		return []*FileInfo{file}, nil
	}

	return []*FileInfo{}, nil
}

// scanDirectory recursively scans a directory for files
func (s *Scanner) scanDirectory(dirPath string) ([]*FileInfo, error) {
	var files []*FileInfo

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if d.IsDir() {
			if s.shouldExcludeDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Process files
		info, err := d.Info()
		if err != nil {
			return err
		}

		file, err := s.scanFile(path, info)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", path, err)
			return nil
		}

		if file != nil {
			files = append(files, file)
		}

		return nil
	})

	return files, err
}

// scanFile processes a single file
func (s *Scanner) scanFile(filePath string, info os.FileInfo) (*FileInfo, error) {
	// Check if file should be excluded
	if s.shouldExcludeFile(filePath) {
		return nil, nil
	}

	// Check file extension
	if !s.hasValidExtension(filePath) {
		return nil, nil
	}

	// Check file size
	if info.Size() > s.options.MaxFileSize {
		return &FileInfo{
			Path:  filePath,
			Size:  info.Size(),
			Type:  s.getFileType(filePath),
			Error: fmt.Sprintf("file too large: %d bytes", info.Size()),
		}, nil
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &FileInfo{
			Path:  filePath,
			Size:  info.Size(),
			Type:  s.getFileType(filePath),
			Error: fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	// Extract text content based on file type
	textContent, err := s.extractTextContent(filePath, content)
	if err != nil {
		return &FileInfo{
			Path:  filePath,
			Size:  info.Size(),
			Type:  s.getFileType(filePath),
			Error: fmt.Sprintf("failed to extract text: %v", err),
		}, nil
	}

	return &FileInfo{
		Path:    filePath,
		Size:    info.Size(),
		Type:    s.getFileType(filePath),
		Content: textContent,
	}, nil
}

// extractTextContent extracts readable text from different file formats
func (s *Scanner) extractTextContent(filePath string, content []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".md", ".markdown":
		return s.extractMarkdownText(string(content)), nil
	case ".html", ".htm":
		return s.extractHTMLText(string(content)), nil
	case ".rst":
		return s.extractRestructuredText(string(content)), nil
	case ".adoc", ".asciidoc":
		return s.extractAsciidocText(string(content)), nil
	default:
		// For plain text and other formats, return as-is
		return string(content), nil
	}
}

// extractMarkdownText extracts text from markdown, removing formatting
func (s *Scanner) extractMarkdownText(content string) string {
	lines := strings.Split(content, "\n")
	var textLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip code blocks (simple detection)
		if strings.HasPrefix(line, "```") {
			continue
		}

		// Skip headers but include the text
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}

		// Skip links but include the text (simple extraction)
		line = s.extractLinkText(line)

		// Skip emphasis markers
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "*", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.ReplaceAll(line, "_", "")
		line = strings.ReplaceAll(line, "`", "")

		if strings.TrimSpace(line) != "" {
			textLines = append(textLines, line)
		}
	}

	return strings.Join(textLines, "\n")
}

// extractHTMLText extracts text from HTML (basic implementation)
func (s *Scanner) extractHTMLText(content string) string {
	// This is a simple implementation - in a real-world scenario,
	// you'd want to use a proper HTML parser

	// Remove script and style tags completely
	content = s.removeHTMLTags(content, "script")
	content = s.removeHTMLTags(content, "style")

	// Remove all HTML tags
	content = s.stripHTMLTags(content)

	// Clean up whitespace
	lines := strings.Split(content, "\n")
	var textLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			textLines = append(textLines, line)
		}
	}

	return strings.Join(textLines, "\n")
}

// extractRestructuredText extracts text from reStructuredText
func (s *Scanner) extractRestructuredText(content string) string {
	lines := strings.Split(content, "\n")
	var textLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip directive lines (start with ..)
		if strings.HasPrefix(line, "..") {
			continue
		}

		// Skip lines that are just underlines/overlines
		if s.isRSTUnderline(line) {
			continue
		}

		textLines = append(textLines, line)
	}

	return strings.Join(textLines, "\n")
}

// extractAsciidocText extracts text from AsciiDoc
func (s *Scanner) extractAsciidocText(content string) string {
	lines := strings.Split(content, "\n")
	var textLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip attribute lines
		if strings.HasPrefix(line, ":") && strings.HasSuffix(line, ":") {
			continue
		}

		// Skip headers but include the text
		if strings.HasPrefix(line, "=") {
			line = strings.TrimSpace(strings.TrimLeft(line, "="))
		}

		if strings.TrimSpace(line) != "" {
			textLines = append(textLines, line)
		}
	}

	return strings.Join(textLines, "\n")
}

// Helper methods

func (s *Scanner) shouldExcludeDir(dirName string) bool {
	for _, excluded := range s.options.ExcludeDirs {
		if dirName == excluded {
			return true
		}
	}
	return false
}

func (s *Scanner) shouldExcludeFile(filePath string) bool {
	fileName := filepath.Base(filePath)
	for _, excluded := range s.options.ExcludeFiles {
		if fileName == excluded {
			return true
		}
	}
	return false
}

func (s *Scanner) hasValidExtension(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, validExt := range s.options.Extensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

func (s *Scanner) getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".md", ".markdown":
		return "markdown"
	case ".html", ".htm":
		return "html"
	case ".rst":
		return "restructuredtext"
	case ".adoc", ".asciidoc":
		return "asciidoc"
	case ".txt", ".text":
		return "text"
	case ".xml":
		return "xml"
	default:
		return "text"
	}
}

func (s *Scanner) extractLinkText(line string) string {
	// Simple markdown link extraction: [text](url) -> text
	for {
		start := strings.Index(line, "[")
		if start == -1 {
			break
		}

		end := strings.Index(line[start:], "](")
		if end == -1 {
			break
		}

		urlEnd := strings.Index(line[start+end+2:], ")")
		if urlEnd == -1 {
			break
		}

		linkText := line[start+1 : start+end]
		line = line[:start] + linkText + line[start+end+2+urlEnd+1:]
	}
	return line
}

func (s *Scanner) removeHTMLTags(content, tagName string) string {
	// Remove opening and closing tags and everything between them
	openTag := "<" + tagName
	closeTag := "</" + tagName + ">"

	for {
		start := strings.Index(strings.ToLower(content), openTag)
		if start == -1 {
			break
		}

		// Find the end of the opening tag
		tagEnd := strings.Index(content[start:], ">")
		if tagEnd == -1 {
			break
		}

		// Find the closing tag
		end := strings.Index(strings.ToLower(content[start:]), closeTag)
		if end == -1 {
			break
		}

		// Remove the entire tag and its content
		content = content[:start] + content[start+end+len(closeTag):]
	}

	return content
}

func (s *Scanner) stripHTMLTags(content string) string {
	// Simple HTML tag removal
	inTag := false
	var result strings.Builder

	for _, char := range content {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func (s *Scanner) isRSTUnderline(line string) bool {
	if len(line) == 0 {
		return false
	}

	// Check if line consists of only underline characters
	underlineChars := "=-~`#\"'^*+<>_"
	firstChar := rune(line[0])

	if !strings.ContainsRune(underlineChars, firstChar) {
		return false
	}

	for _, char := range line {
		if char != firstChar {
			return false
		}
	}

	return len(line) >= 3
}
