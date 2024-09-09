package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	htmlMini "github.com/tdewolff/minify/v2/html"
	"golang.org/x/net/html"

	"gopkg.in/yaml.v2"
)

const imgThumbWidth = 420
const imgJpgQuality = 50

type Site struct {
	Info struct {
		Email   string `yaml:"email"`
		Phone   string `yaml:"phone"`
		Website string `yaml:"website"`
	} `yaml:"info"`
	Sections []struct {
		Title       string `yaml:"title"`
		Class       string `yaml:"class"`
		Content     string `yaml:"content"`
		BreakBefore bool   `yaml:"break_before"`
	} `yaml:"sections"`
}

func main() {

	if len(os.Args) < 2 {
		log.Println(os.Args)
		panic(errors.New("must include path to yaml"))
	}
	pathTo := os.Args[1]

	// parse site.yaml
	siteRaw, err := os.ReadFile(pathTo)
	if err != nil {
		panic(err)
	}
	site := Site{}
	if err := yaml.Unmarshal(siteRaw, &site); err != nil {
		panic(err)
	}

	// load template + generate html
	t := template.New("site.tmpl").Funcs(template.FuncMap{
		"Markdown": func(v string) template.HTML {
			return template.HTML(markdown.ToHTML([]byte(v), nil, nil))
		},
	})
	t = template.Must(t.ParseFiles("data/site.tmpl"))
	buf := bytes.Buffer{}
	if err := t.Execute(&buf, site); err != nil {
		panic(err)
	}

	// crawl html look for images, resize as needed
	outputHtml := buf.String()
	doc, err := html.Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		panic(err)
	}
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			src := ""
			for _, a := range node.Attr {
				if a.Key == "src" {
					src = a.Val
				}
			}
			imgFile, err := os.Open(src)
			if err != nil {
				panic(err)
			}
			img, _, err := image.Decode(imgFile)
			if err != nil {
				panic(err)
			}
			imgBuf := bytes.Buffer{}
			imgType := "jpeg"
			if img.Bounds().Size().X > imgThumbWidth {
				img = resize.Resize(imgThumbWidth, 0, img, resize.Lanczos3)
				img, err = cutter.Crop(img, cutter.Config{
					Width:   16,
					Height:  9,
					Mode:    cutter.Centered,
					Options: cutter.Ratio,
				})
				if err != nil {
					panic(err)
				}
				if err := jpeg.Encode(&imgBuf, img, &jpeg.Options{Quality: imgJpgQuality}); err != nil {
					panic(err)
				}
			} else {
				imgType = "png"
				pngEnc := png.Encoder{CompressionLevel: png.BestCompression}
				if err := pngEnc.Encode(&imgBuf, img); err != nil {
					panic(err)
				}
			}
			b64Img := base64.StdEncoding.EncodeToString(imgBuf.Bytes())
			outputHtml = strings.ReplaceAll(
				outputHtml,
				fmt.Sprintf("img src=\"%s\"", src),
				fmt.Sprintf("img src=\"data:image/%s;base64,%s\"", imgType, b64Img),
			)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)

	m := minify.New()
	m.AddFunc("text/html", htmlMini.Minify)
	m.AddFunc("text/css", css.Minify)
	m.Minify("text/html", os.Stdout, bytes.NewReader([]byte(outputHtml)))

}
