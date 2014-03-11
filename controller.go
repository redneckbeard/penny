package penny

import (
	"github.com/redneckbeard/gadget"
	"github.com/redneckbeard/gadget/env"
	"sort"
)

type PageController struct {
	*gadget.DefaultController
	pages *PageSet
	url   string
}

func NewPageController(paths ...string) gadget.Controller {
	pageset := NewPageSet(env.RelPath(paths...))
	return &PageController{pages: pageset, url: paths[len(paths)-1]}
}

func (c *PageController) IdPattern() string { return `[\w-]+` }
func (c *PageController) Plural() string    { return c.url }

func (c *PageController) Index(r *gadget.Request) (int, interface{}) {
	slice := c.pages.AsSlice()
  filtered := []*Page{}
  for _, p := range slice {
    if p.IsVisible() {
      filtered = append(filtered, p)
    }
  }
	sort.Sort(sort.Reverse(ByDate{filtered}))
	return 200, filtered
}

func (c *PageController) Show(r *gadget.Request) (int, interface{}) {
	if page, ok := c.pages.Get(r.UrlParams["page_id"]); !ok {
		return 404, "Page not found"
	} else {
		if page.IsVisible() || r.Debug() {
			return 200, page
		}
		return 404, "Page not found"
	}
}

func (c *PageController) Feed(r *gadget.Request) (int, interface{}) {
	feed := NewFeed(c.pages)
	feed.Title = "farmr.org"
	feed.Subtitle = "data-driven microfarming"
	response := gadget.NewResponse(feed)
	response.Headers.Set("Content-Type", "application/atom+xml")
	return 200, response
}
