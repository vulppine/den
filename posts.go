package main

import (
	// "bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	// "github.com/vulppine/foxmarks"
	"github.com/vulppine/cmdio-go"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// post
//
// id      - unique ID from SQLite database (only obtained through getPosts())
// Title   - title of post for reference
// Desc    - summary of post
// Date    - Date of post in RFC3339 format (truncd to YYYY-MM-DD during render)
// src     - raw source content as a blob
// Content - rendered content (not stored in database)
// imgs    - the location of images as according to the references in the Markdown document
// loc     - the original location of the post relative to the filesystem
type post struct {
	id      int
	Title   string
	Desc    string
	Date    string
	URL     string
	src     []byte
	Content string
	imgs    []string
	loc     string
}

func newPost(b []byte) *post {
	p := new(post)

	p.src = b

	return p
}

func (p *post) render() *post {
	/* someday, i'll come back for you, foxmarks...
	d := foxmarks.NewDocumentConstructor(bytes.NewReader(p.src)).Parse()

	if d.Content[0].Type == foxmarks.Header1 {
		p.Title = strings.Trim(d.Content[0].Content, "\n ")
		d.Content = d.Content[1:]
	}

	p.Content = d.Render()
	*/

	s := new(strings.Builder)
	md := goldmark.New()
	c := parser.NewContext()
	d := md.Parser().Parse(text.NewReader(p.src), parser.WithContext(c))

	// hacky - maybe find a way to integrate this so we don't have to rely on
	// actually having the raw text segment be the full heading itself,
	// but rather make it tie into the renderer so that we can get the rendered
	// text segment
	if n := d.FirstChild(); n.Kind().String() == "Heading" {
		if e := n.(*ast.Heading); e.Level == 1 {
			l := e.Lines().At(0)
			p.Title = string(l.Value(p.src))
			d.RemoveChild(d, n)
		}
	} else {
		for p.Title == "" {
			p.Title = cmdio.ReadInput("Title needed: please input one now")
		}
	}

	for _, v := range c.References() {
		if r := strings.Split(string(v.Destination()), "/"); r[0] == "img" {
			p.imgs = append(p.imgs, strings.Join(r[1:], "/"))
		}
	}

	md.Renderer().Render(s, p.src, d)
	p.Content = s.String()

	return p
}

func (p *post) writeHTML(o string) error {
	t, err := template.ParseFiles(filepath.Join(templatesrc, "post_template.html"))
	if checkError(err) {
		return err
	}

	f, err := os.Create(o)
	if checkError(err) {
		return err
	}
	err = t.Execute(f, p)
	if checkError(err) {
		return err
	}

	return nil
}

func (p *post) writeSrc(o string) error {
	f, err := os.Create(o)
	if checkError(err) {
		return err
	}

	_, err = f.Write(p.src)
	if checkError(err) {
		return err
	}

	return nil
}

// copyImages copies images from the imgs array to the target directory,
// into a directory named 'img'.
func (p *post) copyImages(d string) error {
	if len(p.imgs) == 0 {
		return nil
	}

	err := os.Mkdir(filepath.Join(d, "img"), 0755)
	if err != nil {
		return err
	}

	for _, v := range p.imgs {
		if f, err := os.Open(filepath.Join(p.loc, v)); err != nil {
			log.Printf("could not access %s, skipping (error: %s)\n", v, err)
		} else {
			if b, err := io.ReadAll(f); err != nil {
				log.Printf("could not access %s, skipping (error: %s)\n", v, err)
			} else {
				i, err := os.Create(filepath.Join(d, "img", v))
				if checkError(err) {
					return err
				}

				_, err = i.Write(b)
				if checkError(err) {
					return err
				}

				i.Close()
			}
		}
	}

	return nil
}

type postListing int

const (
	index   postListing = 0
	rss     postListing = 1
	archive postListing = 2
)

// write index entries to either the index or a RSS file
func writeIndexEntries(p []*post, o string, t postListing) error {
	var m string
	switch t {
	case index:
		m = "index_template.html"
	case rss:
		m = "rss_template.rss"
	case archive:
		m = "archive_template.html"
	}
	e, err := template.ParseFiles(filepath.Join(templatesrc, m))
	if checkError(err) {
		return err
	}

	f, err := os.Create(o)
	if checkError(err) {
		return err
	}
	err = e.Execute(f, p)
	if checkError(err) {
		return err
	}

	return nil
}
