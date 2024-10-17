package render

import (
	"bytes"
	"fmt"
	"io"

	"log"
	"os"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

const pageLayout = `
<!DOCTYPE html>
<html>
	<head>
		{{ .Head }}
	</head>
	<body class="container-lg px-3 my-5 markdown-body">
		{{ .Body }}
	</body>
</html>
`

var (
	head string
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func init() {
	htmlFormatter = html.New(html.WithClasses(true), html.TabWidth(2))
	if htmlFormatter == nil {
		panic("couldn't create html formatter")
	}
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
	if highlightStyle == nil {
		panic(fmt.Sprintf("didn't find style '%s'", styleName))
	}

	// we cache the contets of <head> into a variable
	if _, err := os.Stat(".config/head.html"); err != nil {
		return
	}

	dat, err := os.ReadFile(".config/head.html")
	if err != nil {
		return
	}

	head = string(dat)
}

var pageTemplate, _ = template.New("page").Parse(pageLayout)

// copypasta from https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func renderCode(w io.Writer, codeBlock *ast.CodeBlock, _ bool) {
	defaultLang := ""
	lang := string(codeBlock.Info)
	htmlHighlight(w, string(codeBlock.Literal), lang, defaultLang)
}

func myRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if code, ok := node.(*ast.CodeBlock); ok {
		renderCode(w, code, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func newCustomizedRender() *mdhtml.Renderer {
	opts := mdhtml.RendererOptions{
		Flags:          mdhtml.CommonFlags,
		RenderNodeHook: myRenderHook,
	}
	return mdhtml.NewRenderer(opts)
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	renderer := newCustomizedRender()

	return markdown.Render(doc, renderer)
}

type Page struct {
	Head string
	Body string
}

// takes a bunch of bytes of markdown and renders an html page
func Render(md []byte) []byte {
	page := Page{
		head,
		string(mdToHTML(md)),
	}

	buffer := &bytes.Buffer{}
	pageTemplate.Execute(buffer, page)
	return buffer.Bytes()
}
