package penny

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/redneckbeard/quimby"
	"launchpad.net/goyaml"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	NAV_MODE int = iota
	COMMAND_MODE
	SELECT_MODE
)

var (
	mode, lastMode, selectionStart, selectionEnd int
	command, errMsg                              string
	commands                                     = make(map[string]function)
)

func init() {
	quimby.Add(&NewPost{FrontMatter: &FrontMatter{}}, &Penny{})
}

func switchMode(newMode int) {
	lastMode = mode
	mode = newMode
}

func AddFunctions(functions ...function) {
	for _, f := range functions {
		commands[f.Name()] = f
	}
}

type function interface {
	Name() string
	Run(PageSlice, string) string
}

type NewPost struct {
	*FrontMatter
	*quimby.Flagger
	path string
}

func (c *NewPost) Desc() string {
	return "Generate a markdown file for a new post with front matter."
}

func (c *NewPost) SetFlags() {
	c.StringVar(&c.FrontMatter.Title, "title", "New post", "Title for your new blog post")
	c.StringVar(&c.FrontMatter.Series, "series", "", "Title of the series this post belongs to, if any")
	c.StringVar(&c.FrontMatter.DateString, "date", time.Now().Format(time.UnixDate), "Publication date of your new post, as a UNIX timestamp (i.e. the output of `date`)")
	c.StringVar(&c.path, "dirname", "", "Directory where the new post will be created")

}

func (c *NewPost) Run() {
	if fm, err := goyaml.Marshal(c.FrontMatter); err != nil {
		panic(err)
	} else {
		filename := filepath.Join(c.path, Slugify(c.FrontMatter.Title)+".md")
		if f, err := os.Create(filename); err == nil {
			defer f.Close()
			f.Write(append(fm, []byte("\n---\n")...))
		} else {
			panic(err)
		}
	}
}

type Penny struct {
	*quimby.Flagger
	path string
}

func (c *Penny) Desc() string {
	return "Launch the Penny interface."
}

func (c *Penny) SetFlags() {
	c.StringVar(&c.path, "dirname", "", "Directory containing posts to display stats on")
}

func (c *Penny) Run() {
	var (
		pos    int
		cut    *Page
		logger *log.Logger
	)
	pageset := PageSets[c.path]
	selected := pageset.AsSlice()
	sort.Sort(ByDate{selected})

	if f, err := os.OpenFile("penny.log", os.O_RDWR|os.O_CREATE, 0666); err != nil {
		panic(err)
	} else {
		writer := f
		logger = log.New(writer, "", 0)
	}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawSelected(selected, pos)
	termbox.Flush()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				switchMode(NAV_MODE)
				errMsg = ""
			}
			if ev.Key == termbox.KeyCtrlV && mode == NAV_MODE {
				switchMode(SELECT_MODE)
				selectionStart, selectionEnd = pos, pos
			}
			if mode == COMMAND_MODE {
				errMsg = ""
				switch {
				case ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2:
					command = command[:len(command)-1]
				case ev.Key == termbox.KeyEnter:
					cmdParams := strings.SplitN(command, " ", 2)
					if f, ok := commands[cmdParams[0]]; ok {
						start, end := 0, len(selected)-1
						if lastMode == SELECT_MODE {
							if selectionStart <= selectionEnd {
								start, end = selectionStart, selectionEnd+1
							} else {
								start, end = selectionEnd, selectionStart
							}
						}
						logger.Println(start, end)
						errMsg = f.Run(selected[start:end], cmdParams[1])
						command = ""
						switchMode(NAV_MODE)
					} else {
						errMsg = fmt.Sprintf(`Command "%s" not found`, cmdParams[0])
					}
				default:
					char := string(ev.Ch)
					if ev.Key == termbox.KeySpace {
						char = " "
					}
					command = command + char
				}
			} else {
				if ev.Ch == 100 {
					cut = selected[pos]
					selected = append(selected[:pos], selected[pos+1:]...)
					if pos > len(selected) {
						pos--
					}
				}
				if ev.Ch == 106 {
					pos = (pos + 1) % len(selected)
					selectionEnd = pos
				}
				if ev.Ch == 107 {
					if pos == 0 {
						pos = len(selected) - 1
					} else {
						pos--
					}
					selectionEnd = pos
				}
				if ev.Ch == 112 {
					if cut != nil {
						selected = append(selected[:pos+1], append(PageSlice{cut}, selected[pos+1:]...)...)
						pos++
					}
				}
				if ev.Ch == 58 {
					switchMode(COMMAND_MODE)
				}
				if ev.Key == termbox.KeyEnter {
					defaultEditor := os.ExpandEnv(`${EDITOR}`)
					if defaultEditor == "" {
						defaultEditor = "mvim"
					}
					c := exec.Command(defaultEditor, selected[pos].Path)
					c.Stdout = os.Stdout
					c.Start()
					if err := c.Wait(); err != nil {
						panic(err)
					}
				}
			}
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			drawSelected(selected, pos)
			termbox.Flush()
		case termbox.EventResize:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			drawSelected(selected, pos)
			termbox.Flush()
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func drawSelected(selected PageSlice, pos int) {
	width, _ := termbox.Size()
	print_tb(0, 0, termbox.ColorDefault, termbox.ColorYellow, " ¢ Welcome to Penny ¢ ")
	print_tb(0, 1, termbox.ColorDefault, termbox.ColorDefault, "")

	for i, s := range selected.Display(width) {
		textColor, backgroundColor := termbox.ColorDefault, termbox.ColorDefault
		if mode == SELECT_MODE {
			if (selectionEnd >= selectionStart && i >= selectionStart && i <= selectionEnd) || (selectionEnd < selectionStart && i >= selectionEnd && i <= selectionStart) {
				textColor, backgroundColor = termbox.ColorWhite, termbox.ColorBlack
			}
		}
		if i == pos {
			textColor |= termbox.AttrBold
		}
		print_tb(1, i+2, textColor, backgroundColor, s)
	}
	if mode == COMMAND_MODE {
		cli := " >> " + command + strings.Repeat(" ", width-len(command)-4)
		print_tb(0, len(selected)+3, termbox.ColorWhite, termbox.ColorBlack, cli)
		if len(errMsg) > 0 {
			print_tb(0, len(selected)+4, termbox.ColorWhite, termbox.ColorRed, " "+errMsg+strings.Repeat(" ", width-len(errMsg)-1))
		}
	}
	termbox.Flush()

}

func print_tb(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}
