package preprocess

import (
	"bytes"
	"text/template"
)

func Preprocess(plaintext string) string {
	buf := new(bytes.Buffer)
	t := template.Must(template.New("post").Funcs(helpers).Delims("[[", "]]").Parse(plaintext))
	t.Execute(buf, nil)
	processed := buf.String()
	return processed
}
