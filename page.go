package penny

import (
	"bytes"
	"code.google.com/p/cascadia"
	html5 "code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"github.com/knieriem/markdown"
	"github.com/redneckbeard/gadget/env"
	"github.com/redneckbeard/penny/preprocess"
	html "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	whitespace    = regexp.MustCompile(`\W+`)
	nonWhitespace = regexp.MustCompile(`[^\w\s]`)
	selector      = cascadia.MustCompile("p")
)

type Page struct {
	*FrontMatter
	Body           html.HTML
	Markdown       string
	WordCount      int
	Path           string
	Next, Previous *Page
	Series         *series
}

func (p *Page) IsVisible() bool {
	switch {
	case env.Debug:
  case p.Draft:
    return false
	case time.Now().After(p.FrontMatter.Date):
	default:
		return false
	}
	return true
}

func (p *Page) Save() error {
	f, err := os.Create(p.Path)
	defer f.Close()
	if err != nil {
		return err
	}
	fmt.Fprintf(f, `%s
---
%s`, p.FrontMatter, p.Markdown)
	return nil
}

func (p *Page) SetDescription() {
	if p.Description == "" {
		doc, err := html5.Parse(strings.NewReader(string(p.Body)))
		if err != nil {
			return
		}
		paragraphs := selector.MatchAll(doc)
		if len(paragraphs) > 0 {
			firstParagraph := paragraphs[0]
			for c := firstParagraph.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html5.TextNode {
					p.Description += c.Data
				} else {
					p.Description += c.FirstChild.Data
				}
			}
		}
	}
}

func NewPage(path, src string) (*Page, error) {
	dst := preprocess.Preprocess(src)
	wc := len(strings.Fields(dst))
	dst = fromMarkdown(dst)
	return &Page{Path: path, Markdown: src, Body: html.HTML(dst), WordCount: wc}, nil
}

func fromMarkdown(src string) string {
	// Dance for the markdown library:
	body := []byte(src)
	bodyReader := bytes.NewReader(body)
	markup := new(bytes.Buffer)
	p := markdown.NewParser(&markdown.Extensions{Smart: true, Notes: true})
	p.Markdown(bodyReader, markdown.ToHTML(markup))
	return markup.String()
}

func Parse(filePath string, isIndex bool) (*Page, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("Failed to read from file:" + filePath)
	}
	contents := strings.SplitN(string(raw), "---", 2)
	config := NewFrontMatter(contents[0])
	path, _ := filepath.Abs(filePath)
	page, err := NewPage(path, contents[1])
	if err != nil {
		return nil, err
	}
	page.FrontMatter = config
	page.SetDescription()
	return page, nil
}
