//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_markdown_editor.go; edit that file, not this copy.
// Host logic of examples/markdown_editor/main.cpp: the editor starts with the sample Markdown;
// Clear empties it and Load Sample puts the sample back. After either, the preview shows the
// converted HTML (as text, like the original: "for now, set as text") and #stats counts words
// and characters of the content.
//
// The original never listens to the textarea: typing does not update the preview or the stats.
// Its header checks compare a 3-char prefix with "### " and a 2-char prefix with "## ", so they
// never match; this port does what they mean to (<h3> / <h2>).
package main

import (
	"fmt"
	"regexp"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

const defaultMarkdown = "# Welcome to KUI Markdown Editor\n" +
	"\n" +
	"This is a **live preview** markdown editor built with KUI.\n" +
	"\n" +
	"## Features\n" +
	"\n" +
	"- Real-time markdown to HTML preview\n" +
	"- Dark theme with syntax highlighting\n" +
	"- Word and character counting\n" +
	"- Split pane layout\n" +
	"\n" +
	"## Code Example\n" +
	"\n" +
	"```cpp\n" +
	"#include <kui/kui.h>\n" +
	"\n" +
	"int main() {\n" +
	"    kui::Platform::init();\n" +
	"    auto window = kui::Window::create({.title = \"Hello\"});\n" +
	"    return kui::Platform::run();\n" +
	"}\n" +
	"```\n" +
	"\n" +
	"## Formatting\n" +
	"\n" +
	"You can use **bold**, *italic*, and `inline code`.\n" +
	"\n" +
	"> Blockquotes are also supported for callouts and notes.\n" +
	"\n" +
	"---\n" +
	"\n" +
	"### Lists\n" +
	"\n" +
	"- Item one\n" +
	"- Item two\n" +
	"- Item three\n" +
	"\n" +
	"Enjoy writing with KUI!\n"

var (
	boldRe   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRe = regexp.MustCompile(`\*(.+?)\*`)
	codeRe   = regexp.MustCompile("`(.+?)`")
	escaper  = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
)

// markdownToHTML is the original's basic-subset converter, line by line.
func markdownToHTML(md string) string {
	var html strings.Builder
	inCode := false
	lines := strings.Split(md, "\n")
	if strings.HasSuffix(md, "\n") || md == "" {
		lines = lines[:len(lines)-1] // std::getline yields no line after the final newline
	}
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "```"):
			if inCode {
				html.WriteString("</pre>\n")
			} else {
				html.WriteString("<pre><code>")
			}
			inCode = !inCode
		case inCode:
			html.WriteString(escaper.Replace(line) + "\n")
		case line == "---" || line == "***" || line == "___":
			html.WriteString("<hr>\n")
		case strings.HasPrefix(line, "### "):
			html.WriteString("<h3>" + line[4:] + "</h3>\n")
		case strings.HasPrefix(line, "## "):
			html.WriteString("<h2>" + line[3:] + "</h2>\n")
		case strings.HasPrefix(line, "# "):
			html.WriteString("<h1>" + line[2:] + "</h1>\n")
		case strings.HasPrefix(line, "> "):
			html.WriteString("<blockquote><p>" + line[2:] + "</p></blockquote>\n")
		case strings.HasPrefix(line, "- "):
			html.WriteString("<li>" + line[2:] + "</li>\n")
		case line == "":
			html.WriteString("<br>\n")
		default:
			processed := boldRe.ReplaceAllString(line, "<strong>$1</strong>")
			processed = italicRe.ReplaceAllString(processed, "<em>$1</em>")
			processed = codeRe.ReplaceAllString(processed, "<code>$1</code>")
			html.WriteString("<p>" + processed + "</p>\n")
		}
	}
	if inCode {
		html.WriteString("</pre>\n")
	}
	return html.String()
}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	content := defaultMarkdown
	setText := func(id, text string) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			_ = el.SetTextContent(text)
			el.Close()
		}
	}
	updatePreview := func() {
		preview, err := doc.GetElementByID("preview")
		if err != nil || preview == nil {
			return
		}
		defer preview.Close()
		for {
			child, err := preview.FirstChild()
			if err != nil || child == nil {
				break
			}
			_ = preview.RemoveChild(child)
			child.Close()
		}
		if wrapper, err := doc.CreateElement("div"); err == nil && wrapper != nil {
			_ = wrapper.SetTextContent(markdownToHTML(content))
			_ = preview.AppendChild(wrapper)
			wrapper.Close()
		}
		words := len(strings.FieldsFunc(content, func(r rune) bool {
			return r == ' ' || r == '\n' || r == '\r' || r == '\t'
		}))
		setText("stats", fmt.Sprintf("%d words, %d characters", words, len(content)))
	}

	setText("editor", content)
	for id, text := range map[string]string{"btn-clear": "", "btn-sample": defaultMarkdown} {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		text := text
		if _, err := button.AddEventListener("click", func(*kui.Event) {
			content = text
			setText("editor", content)
			updatePreview()
		}); err != nil {
			return err
		}
	}
	updatePreview()
	return nil
}
