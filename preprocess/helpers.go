package preprocess

import (
	"fmt"
	"github.com/nfnt/resize"
	"github.com/redneckbeard/gadget/env"
	"image"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	html "html/template"
)

var helpers = template.FuncMap{
	"amazon": amazon,
	"image":  magicImage,
}

func amazon(tagName, ASIN string, quantity int) string {
	u, _ := url.Parse(fmt.Sprintf("http://www.amazon.com/dp/%s/", ASIN))
	v := url.Values{}
	v.Set("tag", tagName)
	u.RawQuery = v.Encode()
	return u.String()
}

func magicImage(path string, width, height uint, caption, class string) html.HTML {
	f, err := os.Open(env.RelPath("static", path))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	resizedPath := fmt.Sprintf("%s%dx%d%s", path[:len(path)-len(ext)], width, height, ext)

	finfo, err := os.Stat(env.RelPath("static", resizedPath))
	
	if err != nil || finfo.Size() == 0 {
		fmt.Printf("Resizing image '%s'...", path)
		img, _, err := image.Decode(f)
		if err != nil {
			panic(err)
		}
		resized := resize.Resize(width, height, img, resize.Bicubic)

		resizedFile, err := os.Create(env.RelPath("static", resizedPath))
		if err != nil {
			panic(err)
		}
		defer resizedFile.Close()

		switch ext {
		case ".jpg", ".jpeg":
			jpeg.Encode(resizedFile, resized, nil)
		case ".png":
			png.Encode(resizedFile, resized)
		}
		fmt.Println(" complete.")
	}
	return html.HTML(fmt.Sprintf(`
<div class="%s">
<figure>
	<a href="/static/%s">
		<img src="/static/%s" />
	</a>
	<figcaption>%s</figcaption>
	</figure>
</div>
`, class, path, resizedPath, caption))
}
