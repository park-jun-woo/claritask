package render

import (
	"bytes"
	"fmt"
	htmlpkg "html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

// codeBlockPattern matches markdown code blocks
var codeBlockPattern = regexp.MustCompile("```[\\s\\S]*?```")

// Threshold is the character limit for inline messages vs file
const Threshold = 1000

// ShouldRenderAsFile returns true if content should be sent as HTML file
func ShouldRenderAsFile(markdown string) bool {
	return len(markdown) >= Threshold || codeBlockPattern.MatchString(markdown)
}

// ToTelegramHTML converts markdown to Telegram-compatible HTML (limited tags)
// Supports: <b>, <i>, <u>, <code>, <pre>, <a>
func ToTelegramHTML(markdown string) string {
	// Escape HTML entities first
	text := htmlpkg.EscapeString(markdown)

	// Links: [text](url) - must process before other formatting
	// URL is already escaped, so &amp; may appear - we need to handle that
	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllStringFunc(text, func(match string) string {
		re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			// Unescape the URL (it was escaped earlier)
			url := strings.ReplaceAll(parts[2], "&amp;", "&")
			return fmt.Sprintf(`<a href="%s">%s</a>`, url, parts[1])
		}
		return match
	})

	// Bold: **text** or __text__
	text = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(text, "<b>$1</b>")
	text = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(text, "<b>$1</b>")

	// Italic: *text* or _text_
	text = regexp.MustCompile(`\*([^*]+?)\*`).ReplaceAllString(text, "<i>$1</i>")
	text = regexp.MustCompile(`_([^_]+?)_`).ReplaceAllString(text, "<i>$1</i>")

	// Code block: ```code```
	text = regexp.MustCompile("```([\\s\\S]*?)```").ReplaceAllString(text, "<pre>$1</pre>")

	// Inline code: `code`
	text = regexp.MustCompile("`([^`]+?)`").ReplaceAllString(text, "<code>$1</code>")

	// Clean up any remaining markdown link brackets that weren't matched
	// (e.g., malformed links)
	text = regexp.MustCompile(`\[([^\]]*)\]`).ReplaceAllString(text, "$1")

	return text
}

// ToHTMLFile converts markdown to a complete HTML document
func ToHTMLFile(markdown, title string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithHardWraps(),
			gmhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("convert markdown: %w", err)
	}

	bodyContent := buf.String()

	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            line-height: 1.6;
            padding: 20px;
            max-width: 800px;
            margin: 0 auto;
            color: #24292e;
            background-color: #ffffff;
        }
        h1, h2, h3 { margin-top: 24px; margin-bottom: 16px; font-weight: 600; }
        h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
        h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
        pre {
            background-color: #f6f8fa;
            padding: 16px;
            border-radius: 6px;
            overflow: auto;
            font-size: 85%%;
        }
        code {
            font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
            background-color: rgba(27,31,35,0.05);
            padding: 0.2em 0.4em;
            border-radius: 3px;
            font-size: 85%%;
        }
        pre code {
            background-color: transparent;
            padding: 0;
        }
        table {
            border-collapse: collapse;
            width: 100%%;
            margin-bottom: 16px;
        }
        th, td {
            border: 1px solid #dfe2e5;
            padding: 6px 13px;
        }
        th { background-color: #f6f8fa; }
        blockquote {
            border-left: 0.25em solid #dfe2e5;
            padding: 0 1em;
            color: #6a737d;
            margin: 0;
        }
        ul, ol { padding-left: 2em; }
        li { margin: 0.25em 0; }
        a { color: #0366d6; text-decoration: none; }
        a:hover { text-decoration: underline; }
        hr { border: none; border-top: 1px solid #eaecef; margin: 24px 0; }

        @media (prefers-color-scheme: dark) {
            body { color: #c9d1d9; background-color: #0d1117; }
            h1, h2 { border-bottom-color: #30363d; }
            a { color: #58a6ff; }
            pre { background-color: #161b22; }
            code { background-color: rgba(110,118,129,0.4); }
            th, td { border-color: #30363d; }
            th { background-color: #161b22; }
            blockquote { border-left-color: #30363d; color: #8b949e; }
            hr { border-top-color: #30363d; }
        }
    </style>
</head>
<body>
%s
</body>
</html>`, title, bodyContent)

	return fullHTML, nil
}

// ExtractTitle extracts title from markdown (first # heading or first line)
func ExtractTitle(markdown string) string {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "## ") {
			return strings.TrimPrefix(line, "## ")
		}
	}
	// Fallback: first non-empty line, truncated
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if len(line) > 50 {
				return line[:50] + "..."
			}
			return line
		}
	}
	return "Report"
}
