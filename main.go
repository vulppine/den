package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	// "github.com/vulppine/foxmarks"
)

var wd string
var templatesrc string

func isExist(f string) bool {
	_, err := os.Stat(f)
	return !os.IsNotExist(err)
}

var vb bool

func verbose(i interface{}) {
	if vb {
		log.Println(i)
	}
}

func checkError(e error) bool {
	if e != nil {
		return true
	}

	return false
}

func main() {

	flag.Parse()
	if flag.Arg(0) == "" {
		panic("no args given")
	}

	var err error
	var b *blog
	if !isExist("blog.db") {
		b, err = createBlog("", ReadInput("Please input the base URL of your blog (including any preceding folders)"))
		if checkError(err) {
			panic(err)
		}
	} else {
		b, err = openBlog("")
	}

	switch flag.Arg(0) {
	case "add":
		if flag.Arg(1) == "" {
			panic("missing argument to add")
		}

		y, err := ioutil.ReadFile(flag.Arg(1))
		if checkError(err) {
			panic(err)
		}

		p := newPost(y)
		i, err := filepath.Abs(flag.Arg(1))
		if checkError(err) {
			panic(err)
		}
		p.loc = filepath.Dir(i)

		p, err = b.addPost(p.render())
		if checkError(err) {
			panic(err)
		}
		err = b.addImages(p)
		if checkError(err) {
			panic(err)
		}
		m, err := b.readImages(p)
		if checkError(err) {
			panic(err)
		}
		d := filepath.Join(b.config.rootFolder, "posts", strconv.Itoa(p.id))
		err = os.MkdirAll(d, 0755)
		err = p.writeHTML(filepath.Join(d, "index.html"))
		err = p.writeSrc(filepath.Join(d, "post.md"))
		err = writeImages(m, filepath.Join(d))
		if checkError(err) {
			panic(err)
		}

		p = new(post)

		r, err := b.getPosts(p, 10)
		if checkError(err) {
			panic(err)
		}
		err = writeIndexEntries(r, filepath.Join(b.config.rootFolder, "index.html"), index)
		if checkError(err) {
			panic(err)
		}
		err = writeIndexEntries(r, filepath.Join(b.config.rootFolder, "feed.rss"), rss)
		if checkError(err) {
			panic(err)
		}

		r, err = b.getPosts(p, 0)
		err = writeIndexEntries(r, filepath.Join(b.config.rootFolder, "archive.html"), archive)
		if checkError(err) {
			panic(err)
		}
	case "update":
		if flag.Arg(1) != "" {
			if flag.Arg(2) != "" {
				f, err := ioutil.ReadFile(flag.Arg(2))
				if checkError(err) {
					panic(err)
				}

				p := newPost(f)
				r, err := filepath.Abs(flag.Arg(2))
				if checkError(err) {
					panic(err)
				}
				p.loc = filepath.Dir(r)

				p = p.render()
				p.id, err = strconv.Atoi(flag.Arg(1))
				if checkError(err) {
					panic(err)
				}

				err = b.updatePost(p)
				if checkError(err) {
					panic(err)
				}
				m, err := b.readImages(p)
				if checkError(err) {
					panic(err)
				}

				d := filepath.Join(b.config.rootFolder, "posts", strconv.Itoa(p.id))
				err = p.writeHTML(filepath.Join(d, "index.html"))
				err = p.writeSrc(filepath.Join(d, "post.md"))
				err = writeImages(m, filepath.Join(d))
				if checkError(err) {
					panic(err)
				}

				p = new(post)

				i, err := b.getPosts(p, 10)
				if checkError(err) {
					panic(err)
				}
				err = writeIndexEntries(i, filepath.Join(b.config.rootFolder, "index.html"), index)
				if checkError(err) {
					panic(err)
				}
				err = writeIndexEntries(i, filepath.Join(b.config.rootFolder, "feed.rss"), rss)
				if checkError(err) {
					panic(err)
				}

				i, err = b.getPosts(p, 0)
				err = writeIndexEntries(i, filepath.Join(b.config.rootFolder, "archive.html"), archive)
				if checkError(err) {
					panic(err)
				}

				return
			} else {
				panic("missing argument to update")
			}
		}

		p := new(post)

		if flag.Arg(1) != "" {
			p.id, err = strconv.Atoi(flag.Arg(1))
			if checkError(err) {
				panic(err)
			}
		}

		r, err := b.getPosts(p, 10)
		if checkError(err) {
			panic(err)
		}
		err = writeIndexEntries(r, filepath.Join(b.config.rootFolder, "index.html"), index)
		if checkError(err) {
			panic(err)
		}
		err = writeIndexEntries(r, filepath.Join(b.config.rootFolder, "feed.rss"), rss)
		if checkError(err) {
			panic(err)
		}

		r, err = b.getPosts(p, 0)
		err = writeIndexEntries(r, filepath.Join(b.config.rootFolder, "archive.html"), archive)
		if checkError(err) {
			panic(err)
		}

		for _, i := range r {
			d := filepath.Join(b.config.rootFolder, "posts", strconv.Itoa(i.id))
			i.render()
			err = i.writeHTML(filepath.Join(d, "index.html"))
			err = i.writeSrc(filepath.Join(d, "post.md"))
		}
	case "rm":
		if flag.Arg(1) != "" {
			p := new(post)
			p.id, err = strconv.Atoi(flag.Arg(1))
			if err != nil {
				panic(err)
			}

			err = b.removePost(p)
			if err != nil {
				panic(err)
			}
		} else {
			panic("missing argument to rm")
		}
	}
}

func init() {
	wd, err := os.Getwd()
	if checkError(err) {
		panic(err)
	}

	templatesrc = filepath.Join(wd, "templates")
}
