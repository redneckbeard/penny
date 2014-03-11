package penny

import (
	"encoding/xml"
	"sort"
	"time"
)

type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr,omitempty"`
	Href string `xml:"href,attr"`
}

type Content struct {
	XMLName xml.Name `xml:"content"`
	Type    string   `xml:"type,attr"`
	Text    string   `xml:",chardata"`
}

type Entry struct {
	Title   string `xml:"title"`
	Summary string `xml:"summary,omitempty"`
	Content *Content
	Date    time.Time `xml:"updated"`
	Links   []*Link   `xml:"link"`
	Name    string    `xml:"author>name"`
	Uri   string    `xml:"author>uri"`
	Id      string    `xml:"id"`
}

func NewEntry(p *Page) *Entry {
	entry := &Entry{
		Title:   p.Title,
		Summary: p.Description,
		Date:    p.Date,
		Content: &Content{Type: "html", Text: string(p.Body)},
		Name:    "Farmer Paul",
		Uri:   "http://farmr.org/",
		Id:      "http://farmr.org/" + p.Path,
	}
	entry.Links = []*Link{
		&Link{Rel: "alternate", Type: "text/html", Href: "http://farmr.org/" + p.Path},
	}
	return entry
}

type Feed struct {
	XMLName   xml.Name  `xml:"feed"`
	Links     []*Link   `xml:"link"`
	Namespace string    `xml:"xmlns,attr"`
	Title     string    `xml:"title"`
	Subtitle  string    `xml:"subtitle"`
	Id        string    `xml:"id"`
	Updated   time.Time `xml:"updated"`
	Entries   []*Entry  `xml:"entry"`
}

func NewFeed(pageset *PageSet) *Feed {
	feed := &Feed{
		Namespace: "http://www.w3.org/2005/Atom",
		Id:        "http://farmr.org/blog/feed",
		Links: []*Link{
			&Link{Rel: "alternate", Type: "text/html", Href: "http://farmr.org/blog"},
			&Link{Rel: "self", Type: "application/atom+xml", Href: "http://farmr.org/blog/feed"},
		},
	}

	published := pageset.AsSlice()
	sort.Sort(sort.Reverse(ByDate{published}))
	for _, page := range published {
		if !page.Date.After(time.Now()) {
			feed.Entries = append(feed.Entries, NewEntry(page))
		}
	}
	if len(feed.Entries) > 0 {
		feed.Updated = feed.Entries[0].Date
	}
	return feed
}
