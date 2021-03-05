package main

import (
	"bytes"
	"path/filepath"
	"os"
	"strings"
    "text/template"

	"github.com/vulppine/foxmarks"
)

// post
//
// id      - unique ID from SQLite database (only obtained through getPosts())
// Title   - title of post for reference
// Desc    - summary of post
// Date    - Date of post in RFC3339 format (truncd to YYYY-MM-DD during render)
// src     - raw source content as a blob
// Content - rendered content (not stored in database)
type post struct {
	id int
	Title string
	Desc string
	Date string
	URL string
	src []byte
	Content string
}

func newPost(b []byte) *post {
	p := new(post)

	p.src = b

	return p
}

func (p *post) render() *post {
	d := foxmarks.NewDocumentConstructor(bytes.NewReader(p.src)).Parse()

	if d.Content[0].Type == foxmarks.Header1 {
		p.Title = strings.Trim(d.Content[0].Content, "\n ")
		d.Content = d.Content[1:]
	}

	p.Content = d.Render()

	return p
}

func (p *post) writeHTML(o string) error {
	t, err := template.ParseFiles(filepath.Join(templatesrc, "post_template.html"))
	if checkError(err) { return err }

	f, err := os.Create(o)
	if checkError(err) { return err }
	err = t.Execute(f, p)
	if checkError(err) { return err }

	return nil
}

func (p *post) writeSrc(o string) error {
	f, err := os.Create(o)
	if checkError(err) { return err }

	_, err = f.Write(p.src)
	if checkError(err) { return err }

	return nil
}

type postListing int

const (
	index postListing = 0
	rss postListing = 1
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
	if checkError(err) { return err }

	f, err := os.Create(o)
	if checkError(err) { return err }
	err = e.Execute(f, p)
	if checkError(err) { return err }

	return nil
}
