package penny

import (
	"github.com/redneckbeard/gadget/templates"
	"reflect"
)

func init() {
	templates.AddHelper("isPage", isPage)
}

func isPage(i interface{}) bool {
	return reflect.TypeOf(&Page{}) == reflect.TypeOf(i)
}

