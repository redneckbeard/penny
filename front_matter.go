package penny

import (
	"launchpad.net/goyaml"
	"strings"
	"time"
)

type FrontMatter struct {
	Title               string
	Description string 
	Series string    
	Date                time.Time `yaml:"-"`
	DateString          string    `yaml:"date"`
	Draft               bool      
	Scripts             []string  
	Slug                string    `yaml:"-"`
	Stylesheets         []string  
}

func (f *FrontMatter) String() string {
	f.DateString = f.Date.Format(time.UnixDate)
	s, _ := goyaml.Marshal(f)
	return string(s)
}

func NewFrontMatter(rawData string) *FrontMatter {
	frontMatter := new(FrontMatter)
	if err := goyaml.Unmarshal([]byte(rawData), frontMatter); err != nil {
		panic(rawData)
	}
	frontMatter.Title = strings.Trim(frontMatter.Title, " ")
	frontMatter.Slug = Slugify(frontMatter.Title)
	if date, err := time.Parse(time.UnixDate, frontMatter.DateString); err == nil {
		frontMatter.Date = date
	}
	return frontMatter
}

func Slugify(title string) string {
	slug := nonWhitespace.ReplaceAllString(strings.ToLower(title), " ")
	slug = whitespace.ReplaceAllString(strings.Trim(slug, " "), "-")
	return slug
}
