package views

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/koren/tuimail/internal/models"
)

const maxEmailWidth = 80

// Cyberpunk styles for email reader
var (
	readerColorNeonCyan   = lipgloss.Color("#00FFFF")
	readerColorNeonPink   = lipgloss.Color("#FF2E97")
	readerColorNeonPurple = lipgloss.Color("#BD00FF")
	readerColorTextBright = lipgloss.Color("#EAEAEA")
	readerColorTextNormal = lipgloss.Color("#B8B8B8")
	readerColorTextMuted  = lipgloss.Color("#5C5C7A")
	readerColorTextDim    = lipgloss.Color("#3D3D5C")
	readerColorBgLight    = lipgloss.Color("#16213E")

	readerHeaderStyle = lipgloss.NewStyle().
				Background(readerColorBgLight).
				BorderStyle(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "╰",
			BottomRight: "╯",
		}).
		BorderForeground(readerColorTextDim).
		Padding(1, 2)

	readerLabelStyle = lipgloss.NewStyle().
				Foreground(readerColorNeonPurple).
				Bold(true).
				Width(10)

	readerValueStyle = lipgloss.NewStyle().
				Foreground(readerColorTextBright)

	readerSubjectStyle = lipgloss.NewStyle().
				Foreground(readerColorNeonCyan).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	readerBodyStyle = lipgloss.NewStyle().
			Foreground(readerColorTextNormal)

	readerAttachmentStyle = lipgloss.NewStyle().
				Foreground(readerColorTextMuted).
				MarginTop(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(readerColorTextDim).
				PaddingTop(1)

	readerAttachmentItemStyle = lipgloss.NewStyle().
					Foreground(readerColorTextMuted)

	readerScrollStyle = lipgloss.NewStyle().
				Foreground(readerColorNeonPink).
				Align(lipgloss.Right).
				Bold(true)

	readerQuoteStyle = lipgloss.NewStyle().
				Foreground(readerColorTextMuted).
				PaddingLeft(2).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(readerColorTextDim)
)

// EmailReaderView displays a single email
type EmailReaderView struct {
	email       *models.Email
	width       int
	height      int
	contentWidth int
	scrollY     int
	lines       []string
	renderer    *glamour.TermRenderer
}

// NewEmailReaderView creates a new email reader view
func NewEmailReaderView(email *models.Email, width, height int) *EmailReaderView {
	contentWidth := width
	if contentWidth > maxEmailWidth {
		contentWidth = maxEmailWidth
	}

	// Create glamour renderer for markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(contentWidth-4),
	)

	v := &EmailReaderView{
		email:        email,
		width:        width,
		height:       height,
		contentWidth: contentWidth,
		renderer:     renderer,
	}
	v.prepareContent()
	return v
}

// SetSize updates the view dimensions
func (v *EmailReaderView) SetSize(width, height int) {
	contentWidth := width
	if contentWidth > maxEmailWidth {
		contentWidth = maxEmailWidth
	}

	if v.contentWidth != contentWidth {
		v.contentWidth = contentWidth
		v.width = width
		v.prepareContent()
	}
	v.height = height
}

// ScrollUp scrolls the content up
func (v *EmailReaderView) ScrollUp() {
	if v.scrollY > 0 {
		v.scrollY--
	}
}

// ScrollDown scrolls the content down
func (v *EmailReaderView) ScrollDown() {
	maxScroll := len(v.lines) - v.height + 10
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.scrollY < maxScroll {
		v.scrollY++
	}
}

func (v *EmailReaderView) prepareContent() {
	// Get body content
	body := v.email.TextBody
	if body == "" && v.email.HTMLBody != "" {
		body = v.htmlToText(v.email.HTMLBody)
	}
	if body == "" {
		body = v.email.Preview
	}

	// Try to render as markdown if it looks like markdown
	if v.looksLikeMarkdown(body) && v.renderer != nil {
		rendered, err := v.renderer.Render(body)
		if err == nil {
			body = rendered
		}
	}

	// Collapse multiple empty lines (more than 1) into single empty line
	body = regexp.MustCompile(`\n{3,}`).ReplaceAllString(body, "\n\n")
	body = strings.TrimSpace(body)

	// Wrap text to content width
	v.lines = v.wrapText(body, v.contentWidth-4)

	// Remove consecutive empty lines from the result
	v.lines = v.collapseEmptyLines(v.lines)
}

// collapseEmptyLines removes consecutive empty lines, keeping only one
func (v *EmailReaderView) collapseEmptyLines(lines []string) []string {
	var result []string
	prevEmpty := false
	for _, line := range lines {
		isEmpty := strings.TrimSpace(line) == ""
		if isEmpty && prevEmpty {
			continue // Skip consecutive empty lines
		}
		result = append(result, line)
		prevEmpty = isEmpty
	}
	return result
}

// looksLikeMarkdown checks if text appears to be markdown
func (v *EmailReaderView) looksLikeMarkdown(text string) bool {
	mdPatterns := []string{
		`^#+ `,           // Headers
		`\*\*[^*]+\*\*`,  // Bold
		`\*[^*]+\*`,      // Italic
		`\[[^\]]+\]\([^)]+\)`, // Links
		"^```",           // Code blocks
		`^- `,            // Lists
		`^\d+\. `,        // Numbered lists
	}

	for _, pattern := range mdPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// View renders the email
func (v *EmailReaderView) View() string {
	if v.email == nil {
		return lipgloss.NewStyle().
			Foreground(readerColorTextMuted).
			Render("◇ No email selected")
	}

	var b strings.Builder

	// Header section with neon border
	header := v.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Subject with neon styling
	subject := v.email.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	b.WriteString(readerSubjectStyle.Render("◈ " + subject))
	b.WriteString("\n\n")

	// Body with scrolling
	bodyHeight := v.height - 12
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	endIdx := v.scrollY + bodyHeight
	if endIdx > len(v.lines) {
		endIdx = len(v.lines)
	}

	startIdx := v.scrollY
	if startIdx > len(v.lines) {
		startIdx = len(v.lines)
	}

	if startIdx < endIdx {
		visibleLines := v.lines[startIdx:endIdx]

		// Style quoted lines differently
		styledLines := make([]string, len(visibleLines))
		for i, line := range visibleLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, ">") {
				styledLines[i] = readerQuoteStyle.Render(line)
			} else {
				styledLines[i] = readerBodyStyle.Render(line)
			}
		}

		b.WriteString(strings.Join(styledLines, "\n"))
	}

	// Scroll indicator
	if len(v.lines) > bodyHeight {
		percent := 0
		maxScroll := len(v.lines) - bodyHeight
		if maxScroll > 0 {
			percent = (v.scrollY * 100) / maxScroll
		}

		indicator := fmt.Sprintf("▾ %d%% ", percent)
		b.WriteString("\n")
		b.WriteString(readerScrollStyle.Width(v.contentWidth - 4).Render(indicator))
	}

	// Attachments
	if len(v.email.Attachments) > 0 {
		b.WriteString("\n")
		b.WriteString(v.renderAttachments())
	}

	// Center the content if wider than maxEmailWidth
	content := b.String()
	if v.width > v.contentWidth {
		return lipgloss.Place(v.width, 0, lipgloss.Center, lipgloss.Top, content)
	}
	return content
}

func (v *EmailReaderView) renderHeader() string {
	var lines []string

	// From
	if len(v.email.From) > 0 {
		from := v.formatAddresses(v.email.From)
		lines = append(lines,
			readerLabelStyle.Render("▸ From")+
				readerValueStyle.Render(from))
	}

	// To
	if len(v.email.To) > 0 {
		to := v.formatAddresses(v.email.To)
		lines = append(lines,
			readerLabelStyle.Render("▸ To")+
				readerValueStyle.Render(to))
	}

	// CC
	if len(v.email.CC) > 0 {
		cc := v.formatAddresses(v.email.CC)
		lines = append(lines,
			readerLabelStyle.Render("▸ CC")+
				readerValueStyle.Render(cc))
	}

	// Date
	date := v.email.ReceivedAt.Format("Mon, Jan 2, 2006 at 3:04 PM")
	lines = append(lines,
		readerLabelStyle.Render("▸ Date")+
			readerValueStyle.Render(date))

	headerWidth := v.contentWidth - 4
	if headerWidth < 40 {
		headerWidth = 40
	}
	return readerHeaderStyle.Width(headerWidth).Render(strings.Join(lines, "\n"))
}

func (v *EmailReaderView) formatAddresses(addrs []models.EmailAddress) string {
	var parts []string
	for _, addr := range addrs {
		parts = append(parts, addr.String())
	}
	return strings.Join(parts, ", ")
}

func (v *EmailReaderView) renderAttachments() string {
	title := lipgloss.NewStyle().
		Foreground(readerColorNeonPurple).
		Bold(true).
		Render("◈ ATTACHMENTS")

	var items []string
	for _, att := range v.email.Attachments {
		if att.IsInline {
			continue
		}
		size := v.formatSize(att.Size)
		item := readerAttachmentItemStyle.Render(
			fmt.Sprintf("  ◇ %s (%s)", att.Name, size))
		items = append(items, item)
	}

	if len(items) == 0 {
		return ""
	}

	content := title + "\n" + strings.Join(items, "\n")
	return readerAttachmentStyle.Render(content)
}

func (v *EmailReaderView) formatSize(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
	)
	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (v *EmailReaderView) wrapText(text string, width int) []string {
	var lines []string

	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		// Handle pre-wrapped lines (from glamour)
		if len(para) <= width {
			lines = append(lines, para)
			continue
		}

		// Wrap long lines
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}

	return lines
}

func (v *EmailReaderView) htmlToText(html string) string {
	text := html

	// Remove style and script tags with content
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	text = styleRe.ReplaceAllString(text, "")
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	text = scriptRe.ReplaceAllString(text, "")

	// Convert headers to markdown
	for i := 6; i >= 1; i-- {
		headerRe := regexp.MustCompile(fmt.Sprintf(`(?i)<h%d[^>]*>([^<]*)</h%d>`, i, i))
		text = headerRe.ReplaceAllString(text, strings.Repeat("#", i)+" $1\n\n")
	}

	// Convert bold/strong to markdown
	boldRe := regexp.MustCompile(`(?i)<(b|strong)[^>]*>([^<]*)</(b|strong)>`)
	text = boldRe.ReplaceAllString(text, "**$2**")

	// Convert italic/em to markdown
	italicRe := regexp.MustCompile(`(?i)<(i|em)[^>]*>([^<]*)</(i|em)>`)
	text = italicRe.ReplaceAllString(text, "*$2*")

	// Convert links to markdown
	linkRe := regexp.MustCompile(`(?i)<a[^>]+href=["']([^"']+)["'][^>]*>([^<]+)</a>`)
	text = linkRe.ReplaceAllString(text, "[$2]($1)")

	// Convert lists
	text = regexp.MustCompile(`(?i)<li[^>]*>`).ReplaceAllString(text, "- ")
	text = regexp.MustCompile(`(?i)</li>`).ReplaceAllString(text, "\n")

	// Convert paragraphs and breaks
	text = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</tr>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`(?i)<p[^>]*>|<div[^>]*>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`(?i)</td>`).ReplaceAllString(text, "\t")

	// Convert blockquotes
	blockquoteRe := regexp.MustCompile(`(?is)<blockquote[^>]*>(.*?)</blockquote>`)
	text = blockquoteRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := blockquoteRe.FindStringSubmatch(match)
		if len(inner) > 1 {
			lines := strings.Split(inner[1], "\n")
			for i, line := range lines {
				lines[i] = "> " + strings.TrimSpace(line)
			}
			return strings.Join(lines, "\n") + "\n"
		}
		return match
	})

	// Remove remaining tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	text = tagRe.ReplaceAllString(text, "")

	// Decode common HTML entities
	entities := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  "\"",
		"&#39;":   "'",
		"&apos;":  "'",
		"&ndash;": "–",
		"&mdash;": "—",
		"&bull;":  "•",
		"&copy;":  "©",
		"&reg;":   "®",
		"&trade;": "™",
	}
	for entity, char := range entities {
		text = strings.ReplaceAll(text, entity, char)
	}

	// Decode numeric entities
	numEntityRe := regexp.MustCompile(`&#(\d+);`)
	text = numEntityRe.ReplaceAllStringFunc(text, func(match string) string {
		var num int
		fmt.Sscanf(match, "&#%d;", &num)
		if num > 0 && num < 128 {
			return string(rune(num))
		}
		return match
	})

	// Clean up whitespace
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)

	return text
}
