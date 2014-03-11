package penny

import (
	"bytes"
	"fmt"
	"github.com/redneckbeard/gadget/sitemap"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

var PageSets = make(map[string]*PageSet)

type PageSlice []*Page

func (p PageSlice) Len() int      { return len(p) }
func (p PageSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type ByDate struct{ PageSlice }

func (s ByDate) Less(i, j int) bool {
	return s.PageSlice[j].Date.After(s.PageSlice[i].Date)
}

type ByWordCount struct{ PageSlice }

func (s ByWordCount) Less(i, j int) bool {
	return s.PageSlice[j].WordCount > s.PageSlice[i].WordCount
}

func (p PageSlice) Display(width int) (final []string) {
	buf := new(bytes.Buffer)
	rightCols := tabwriter.NewWriter(buf, 5, 1, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	for _, page := range p {
		fmt.Fprintf(rightCols, "%d \t %s\n", page.WordCount, page.Date.Format("02 Jan 06"))
	}
	rightCols.Flush()
	line, err := buf.ReadString('\n')
	i := 0
	rightColsWidth := len(line)
	widthForTitle := width - rightColsWidth - 1
	for err == nil {
		title := p[i].Title
		if len(title) > widthForTitle {
			title = title[:widthForTitle-3] + "..."
		}
		spaces := strings.Repeat(" ", width-len(title)-rightColsWidth)
		final = append(final, title+spaces+line)
		line, err = buf.ReadString('\n')
		i++
	}
	return
}

func (p PageSlice) Link() {
	length := len(p)
	for i := range p {
		page := p[i]
		if i > 0 {
			page.Previous = p[i-1]
		}
		if i < length-1 {
			page.Next = p[i+1]
		}
	}
}

func (p PageSlice) MakeSeries() {
	seriesSlice := []*series{}
	for _, page := range p {
		if title := page.FrontMatter.Series; title != "" {
			titleSeries := &series{Title: title}
			needsAppend := true
			for _, s := range seriesSlice {
				if s.Title == title {
					titleSeries = s
					needsAppend = false
				}
			}
			if needsAppend {
				seriesSlice = append(seriesSlice, titleSeries)
			}
			titleSeries.PageSlice = append(titleSeries.PageSlice, page)
			page.Series = titleSeries
			sort.Sort(ByDate{titleSeries.PageSlice})
		}
	}
}

type series struct {
	PageSlice
	Title string
}

func (s *series) Position(p *Page) int {
	for i, page := range s.PageSlice {
		if p == page {
			return i + 1
		}
	}
	return -1
}

func (s *series) Previous(p *Page) *Page {
	for i, page := range s.PageSlice {
		if p == page && i > 0 {
			return s.PageSlice[i-1]
		}
	}
	return nil
}

func (s *series) Next(p *Page) *Page {
	for i, page := range s.PageSlice {
		if p == page && i < len(s.PageSlice)-1 {
			return s.PageSlice[i+1]
		}
	}
	return nil
}

func (s *series) Len() int {
	return len(s.PageSlice)
}

type PageSet struct {
	pages   map[string]*Page `yaml:"-"`
	Path    string           `yaml:"-"`
	Sitemap bool             `yaml:"sitemap"`
}

func (p *PageSet) Get(slug string) (*Page, bool) {
	page, found := p.pages[slug]
	return page, found
}

func (p *PageSet) Add(slug string, page *Page) {
	p.pages[slug] = page
}

func (ps *PageSet) AsSlice() PageSlice {
	p := PageSlice{}
	for _, v := range ps.pages {
		p = append(p, v)
	}
	return p
}

func (ps *PageSet) Link() {
	p := ps.AsSlice()
	sort.Sort(ByDate{p})
	p.Link()
	p.MakeSeries()
}

func (ps *PageSet) AddSitemap() {
	var urls []*sitemap.Url
	for slug, p := range ps.pages {
		if p.IsVisible() {
			urls = append(urls, &sitemap.Url{
				Loc:     filepath.Join(ps.Path, slug),
				LastMod: p.Date,
			})
		}
	}
	sitemap.Add(ps.Path, urls...)

}

func (ps *PageSet) Configure(path string) {
	if raw, err := ioutil.ReadFile(path); err == nil {
		if err := goyaml.Unmarshal(raw, ps); err != nil {
			panic(raw)
		}
	}
}

func NewPageSet(dir string) *PageSet {
	p := &PageSet{Sitemap: true}
	p.Path = filepath.Base(dir)
	p.pages = make(map[string]*Page)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() {
				switch {
				case filepath.Ext(path) == ".md":
					if page, err := Parse(path, false); err != nil {
						panic(err)
					} else {
						p.Add(page.FrontMatter.Slug, page)
					}
				case filepath.Base(path) == "penny.yml":
					p.Configure(path)
				}
			}
		}
		return err
	})
	p.Link()
	PageSets[dir] = p
	return p
}
