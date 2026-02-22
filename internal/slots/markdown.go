package slots

import (
	"bytes"
	"context"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func RegisterMarkdownSlots(eng *engine.Engine) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// MARKDOWN.RENDER
	eng.Register("markdown.render", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		input := coerce.ToString(resolveValue(node.Value, scope))
		target := "html_content"

		for _, c := range node.Children {
			if c.Name == "content" || c.Name == "text" {
				input = coerce.ToString(resolveValue(c.Value, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		var buf bytes.Buffer
		if err := md.Convert([]byte(input), &buf); err != nil {
			return err
		}

		// Use SafeString to prevent double escaping in Blade
		scope.Set(target, SafeString(buf.String()))
		return nil
	}, engine.SlotMeta{
		Description: "Render Markdown text to HTML.",
		Group:       "Utils",
		Example:     "markdown.render: $md_text { as: $html }",
	})
}
