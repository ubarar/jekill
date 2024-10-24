package render

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

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

type Renderer struct {
	Head           string
	HtmlFormatter  *html.Formatter
	HighlightStyle *chroma.Style
}

func NewRenderer(path string) *Renderer {
	renderer := Renderer{}

	renderer.HtmlFormatter = html.New(html.WithClasses(true), html.TabWidth(2))
	if renderer.HtmlFormatter == nil {
		panic("couldn't create html formatter")
	}
	renderer.HighlightStyle = styles.Get("monokailight")
	if renderer.HighlightStyle == nil {
		panic(fmt.Sprintf("didn't find style"))
	}

	dat, err := os.ReadFile(filepath.Join(path, ".config/head.html"))
	if err == nil {
		renderer.Head = string(dat)
	}

	return &renderer
}

var pageTemplate, _ = template.New("page").Parse(pageLayout)

// copypasta from https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func (r *Renderer) HtmlHighlight(w io.Writer, source, lang, defaultLang string) error {
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
	return r.HtmlFormatter.Format(w, r.HighlightStyle, it)
}

func (r *Renderer) RenderCode(w io.Writer, codeBlock *ast.CodeBlock, _ bool) {
	defaultLang := ""
	lang := string(codeBlock.Info)
	r.HtmlHighlight(w, string(codeBlock.Literal), lang, defaultLang)
}

func (r *Renderer) MyRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if code, ok := node.(*ast.CodeBlock); ok {
		r.RenderCode(w, code, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func (r *Renderer) NewCustomizedRender() *mdhtml.Renderer {
	opts := mdhtml.RendererOptions{
		Flags:          mdhtml.CommonFlags,
		RenderNodeHook: r.MyRenderHook,
	}
	return mdhtml.NewRenderer(opts)
}

func (r *Renderer) MdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	renderer := r.NewCustomizedRender()

	return markdown.Render(doc, renderer)
}

type Page struct {
	Head string
	Body string
}

// takes a bunch of bytes of markdown and renders an html page
func (r *Renderer) Render(md []byte) []byte {
	page := Page{
		r.Head,
		string(r.MdToHTML(md)),
	}

	buffer := &bytes.Buffer{}
	pageTemplate.Execute(buffer, page)
	return buffer.Bytes()
}
