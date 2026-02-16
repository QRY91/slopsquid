package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Page represents a crawled page with extracted text
type Page struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Text       string `json:"text,omitempty"`
	rawHTML    string // unexported, used for link extraction
	Error      string `json:"error,omitempty"`
}

// Options configures crawler behavior
type Options struct {
	MaxDepth    int
	MaxPages    int
	Concurrency int
	Delay       time.Duration
	UserAgent   string
	Verbose     bool
}

// Crawler fetches pages from a website and extracts text
type Crawler struct {
	opts       Options
	base       *url.URL
	visited    map[string]bool
	mu         sync.Mutex
	client     *http.Client
	disallowed []string // robots.txt disallow rules
}

var linkRegex = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'#]+)["']`)

// New creates a crawler rooted at the given URL
func New(rootURL string, opts Options) (*Crawler, error) {
	u, err := url.Parse(rootURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	if opts.MaxDepth == 0 {
		opts.MaxDepth = 10
	}
	if opts.MaxPages == 0 {
		opts.MaxPages = 500
	}
	if opts.Concurrency == 0 {
		opts.Concurrency = 3
	}
	if opts.Delay == 0 {
		opts.Delay = 200 * time.Millisecond
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "SlopSquid/0.3"
	}

	c := &Crawler{
		opts:    opts,
		base:    u,
		visited: make(map[string]bool),
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}

	// Fetch and parse robots.txt
	c.disallowed = c.fetchRobotsTxt()

	return c, nil
}

// Crawl starts from the root URL and returns all discovered pages
func (c *Crawler) Crawl(progress func(n int, url string)) ([]*Page, error) {
	var pages []*Page
	var pagesMu sync.Mutex

	type task struct {
		url   string
		depth int
	}

	queue := make(chan task, c.opts.MaxPages)

	// Seed from sitemap if available
	sitemapURLs := c.fetchSitemap()
	for _, u := range sitemapURLs {
		if !c.isDisallowed(u) {
			if c.tryVisit(u) {
				queue <- task{url: u, depth: 0}
			}
		}
	}

	// Always include the root
	if c.tryVisit(c.base.String()) {
		queue <- task{url: c.base.String(), depth: 0}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, c.opts.Concurrency)

	active := len(c.visited) // track pending work

	for active > 0 {
		select {
		case t := <-queue:
			sem <- struct{}{}
			wg.Add(1)
			go func(t task) {
				defer wg.Done()
				defer func() { <-sem }()

				if c.opts.Delay > 0 {
					time.Sleep(c.opts.Delay)
				}

				page := c.fetch(t.url)

				pagesMu.Lock()
				pages = append(pages, page)
				count := len(pages)
				pagesMu.Unlock()

				if progress != nil {
					progress(count, t.url)
				}

				// Extract and queue links if we haven't hit depth limit
				if page.Error == "" && t.depth < c.opts.MaxDepth {
					links := c.extractLinks(page.rawHTML, t.url)
					for _, link := range links {
						if c.isDisallowed(link) {
							continue
						}
						if c.tryVisit(link) {
							pagesMu.Lock()
							n := len(pages)
							pagesMu.Unlock()
							if n < c.opts.MaxPages {
								c.mu.Lock()
								active++
								c.mu.Unlock()
								queue <- task{url: link, depth: t.depth + 1}
							}
						}
					}
				}

				c.mu.Lock()
				active--
				c.mu.Unlock()
			}(t)
		default:
			// Let goroutines finish
			time.Sleep(50 * time.Millisecond)
			c.mu.Lock()
			a := active
			c.mu.Unlock()
			if a == 0 {
				break
			}
		}

		c.mu.Lock()
		a := active
		c.mu.Unlock()
		if a == 0 {
			break
		}
	}

	wg.Wait()

	return pages, nil
}

func (c *Crawler) fetch(pageURL string) *Page {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return &Page{URL: pageURL, Error: err.Error()}
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)
	req.Header.Set("Accept", "text/html")

	resp, err := c.client.Do(req)
	if err != nil {
		return &Page{URL: pageURL, Error: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &Page{URL: pageURL, StatusCode: resp.StatusCode, Error: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "text/plain") {
		return &Page{URL: pageURL, StatusCode: resp.StatusCode, Error: "not text content"}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return &Page{URL: pageURL, StatusCode: resp.StatusCode, Error: err.Error()}
	}

	raw := string(body)
	text := ExtractText(raw)

	return &Page{
		URL:        pageURL,
		StatusCode: resp.StatusCode,
		Text:       text,
		rawHTML:    raw,
	}
}

// extractLinks finds same-domain links in HTML content
func (c *Crawler) extractLinks(html string, pageURL string) []string {
	pageU, _ := url.Parse(pageURL)
	matches := linkRegex.FindAllStringSubmatch(html, -1)

	var links []string
	seen := make(map[string]bool)

	for _, m := range matches {
		href := m[1]

		// Resolve relative URLs
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		resolved := pageU.ResolveReference(u)

		// Same domain only
		if resolved.Host != c.base.Host {
			continue
		}

		// Only HTTP(S)
		if resolved.Scheme != "http" && resolved.Scheme != "https" {
			continue
		}

		// Strip fragment
		resolved.Fragment = ""

		// Skip non-page resources
		path := strings.ToLower(resolved.Path)
		if isAssetPath(path) {
			continue
		}

		canonical := resolved.String()
		if !seen[canonical] {
			seen[canonical] = true
			links = append(links, canonical)
		}
	}

	return links
}

func isAssetPath(path string) bool {
	exts := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot", ".mp3", ".mp4", ".webm", ".webp",
		".pdf", ".zip", ".tar", ".gz", ".xml", ".json", ".rss", ".atom"}
	for _, ext := range exts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// robots.txt support

func (c *Crawler) fetchRobotsTxt() []string {
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", c.base.Scheme, c.base.Host)

	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil
	}

	return parseRobotsTxt(string(body), c.opts.UserAgent)
}

// parseRobotsTxt extracts Disallow rules that apply to our user agent
func parseRobotsTxt(content string, userAgent string) []string {
	var disallowed []string
	lines := strings.Split(content, "\n")

	// Two-pass: first look for our specific user-agent block,
	// then fall back to * block
	uaLower := strings.ToLower(userAgent)
	var inOurBlock, inStarBlock bool
	var ourRules, starRules []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Strip comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)

		if strings.HasPrefix(lower, "user-agent:") {
			agent := strings.TrimSpace(line[len("user-agent:"):])
			agentLower := strings.ToLower(agent)

			// Reset block tracking on new user-agent
			inOurBlock = strings.Contains(uaLower, agentLower) || strings.Contains(agentLower, "slopsquid")
			inStarBlock = agent == "*"
			continue
		}

		if strings.HasPrefix(lower, "disallow:") {
			path := strings.TrimSpace(line[len("disallow:"):])
			if path == "" {
				continue
			}
			if inOurBlock {
				ourRules = append(ourRules, path)
			} else if inStarBlock {
				starRules = append(starRules, path)
			}
		}
	}

	// Prefer specific rules; fall back to wildcard
	if len(ourRules) > 0 {
		disallowed = ourRules
	} else {
		disallowed = starRules
	}

	return disallowed
}

// isDisallowed checks if a URL path is blocked by robots.txt
func (c *Crawler) isDisallowed(rawURL string) bool {
	if len(c.disallowed) == 0 {
		return false
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	path := u.Path

	for _, rule := range c.disallowed {
		// Exact prefix match (standard robots.txt behavior)
		if strings.HasPrefix(path, rule) {
			return true
		}
		// Wildcard support (e.g., "/*.pdf$")
		if strings.Contains(rule, "*") {
			pattern := strings.ReplaceAll(regexp.QuoteMeta(rule), `\*`, `.*`)
			if strings.HasSuffix(pattern, `\$`) {
				pattern = strings.TrimSuffix(pattern, `\$`) + "$"
			}
			if re, err := regexp.Compile(pattern); err == nil && re.MatchString(path) {
				return true
			}
		}
	}

	return false
}

// sitemap support

var locRegex = regexp.MustCompile(`<loc>([^<]+)</loc>`)

func (c *Crawler) fetchSitemap() []string {
	sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", c.base.Scheme, c.base.Host)

	req, err := http.NewRequest("GET", sitemapURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil
	}

	matches := locRegex.FindAllStringSubmatch(string(body), -1)
	var urls []string
	for _, m := range matches {
		u := strings.TrimSpace(m[1])
		if u != "" && !isAssetPath(strings.ToLower(u)) {
			urls = append(urls, u)
		}
	}

	return urls
}

// visited tracking

func (c *Crawler) markVisited(u string) {
	c.mu.Lock()
	c.visited[u] = true
	c.mu.Unlock()
}

func (c *Crawler) tryVisit(u string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.visited[u] {
		return false
	}
	c.visited[u] = true
	return true
}

// ExtractText strips HTML tags and returns readable text.
// Exported so the CLI can use it for local HTML files too.
func ExtractText(html string) string {
	// Remove script/style blocks
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = scriptRe.ReplaceAllString(html, "")
	html = styleRe.ReplaceAllString(html, "")

	// Remove nav, header, footer — focus on article content
	navRe := regexp.MustCompile(`(?is)<(?:nav|header|footer)[^>]*>.*?</(?:nav|header|footer)>`)
	html = navRe.ReplaceAllString(html, "")

	// Strip remaining tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	text := tagRe.ReplaceAllString(html, " ")

	// Decode common entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&mdash;", "—")
	text = strings.ReplaceAll(text, "&ndash;", "–")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&rarr;", "→")
	text = strings.ReplaceAll(text, "&larr;", "←")

	// Collapse whitespace
	spaceRe := regexp.MustCompile(`[ \t]+`)
	text = spaceRe.ReplaceAllString(text, " ")

	// Clean up blank lines
	lines := strings.Split(text, "\n")
	var clean []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			clean = append(clean, line)
		}
	}

	return strings.Join(clean, "\n")
}
