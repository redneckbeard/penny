package penny

import (
	"os"
)

func init() {
	AddFunctions(&Quit{})
}

type Quit struct {}

func (q *Quit) Name() string { return "quit" }

func (q *Quit) Run(slice PageSlice, rawParams string) string {
	os.Exit(0)
	return ""
}
