package utils

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// StripHTML removes HTML tags and decodes HTML entities from a string
func StripHTML(htmlContent string) string {
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent // Return original content if parsing fails
	}

	// Extract text
	var buf bytes.Buffer
	extractText(doc, &buf)

	// Clean up the text
	text := buf.String()
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n") // Remove excessive newlines
	text = strings.ReplaceAll(text, "  ", " ")        // Remove excessive spaces

	return text
}

// extractText recursively extracts text from HTML nodes
func extractText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Skip script and style elements
		if c.Type == html.ElementNode && (c.Data == "script" || c.Data == "style") {
			continue
		}
		extractText(c, buf)
	}
}

// ExtractFirstParagraph extracts the first paragraph from HTML content
func ExtractFirstParagraph(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return StripHTML(htmlContent) // Fallback to full strip if parsing fails
	}

	var buf bytes.Buffer
	found := false

	var extractFirstParagraph func(*html.Node)
	extractFirstParagraph = func(n *html.Node) {
		if found {
			return
		}

		if n.Type == html.ElementNode && n.Data == "p" {
			found = true
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					buf.WriteString(c.Data)
				}
			}
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractFirstParagraph(c)
		}
	}

	extractFirstParagraph(doc)
	text := buf.String()
	if text == "" {
		// If no paragraph found, return stripped content
		return StripHTML(htmlContent)
	}
	return strings.TrimSpace(text)
} 