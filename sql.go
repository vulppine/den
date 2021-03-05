package main

import (
	"fmt"
	"database/sql"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// blog table schema:
// posts
// - id (int, primary)
// - title (text)
// 	 - blog post titles
// - desc (text)
//   - blog post descriptions
// - date_added (date)
//   - date of post when it was added, excluding time
// - src (blob)
//   - the source of the post (primarily in markdown)

type blog struct {
	db *sql.DB
	config struct {
		hostURL    string // the host URL, including preceding folders
		rootFolder string // the root folder, where den folders are stored
	}
}

func openBlog(d string) (*blog, error) {
	var err error
	b := new(blog)

	if d == "" { d = wd }

	b.db, err = sql.Open("sqlite3", filepath.Join(d, "blog.db"))
	if checkError(err) {
		return nil, err
	}

	res, err := b.db.Query("SELECT value FROM config WHERE option = 'rootfolder'")
	if !res.Next() || checkError(err) {
		return nil, fmt.Errorf("could not get root folder, aborting")
	}
	res.Scan(&b.config.rootFolder)
	res.Close()

	res, err = b.db.Query("SELECT value FROM config WHERE option = 'hosturl'")
	if res.Next() {
		res.Scan(&b.config.hostURL)
	} else if checkError(err) {
		return nil, fmt.Errorf("an error occurred during sql transaction")
	}
	res.Close()

	return b, nil
}

func createBlog(d string, u string) (*blog, error) {
	var err error
	b := new(blog)

	if d == "" { d = wd }

	if e := isExist(filepath.Join(d, "blog.db")); e {
		return nil, fmt.Errorf("blog database exists")
	}

	b.db, err = sql.Open("sqlite3", filepath.Join(d, "blog.db"))
	if checkError(err) { return nil, err }

	_, err = b.db.Exec("CREATE TABLE posts (id INTEGER NOT NULL PRIMARY KEY, title TEXT, desc TEXT, date_added DATE, src BLOB)")
	_, err = b.db.Exec("CREATE TABLE config (option TEXT, value TEXT)")
	_, err = b.db.Exec("INSERT INTO config (option, value) VALUES ('rootfolder', 'blog'), ('hosturl', ?)", u)

	return b, err
}

func (b *blog) addPost(p *post) (*post, error) {
	/*
	res, err := b.Query("SELECT title FROM posts WHERE title = ?", p.Title)
	if res.Next() || checkError(err) {
		return nil, fmt.Errorf("attempted to add post with existing title or an error occurred - perhaps update?")
	}
	*/
	p.Date = time.Now().Format("2006-01-02")

	i, err := b.db.Exec(
		"INSERT INTO posts (title, desc, date_added, src) VALUES (?, ?, ?, ?)",
		p.Title, p.Desc, p.Date, p.src,
	)
	if checkError(err) { return nil, err }
	d, err := i.LastInsertId()
	p.id = int(d)

	/*
	r, err := b.db.Query("SELECT id FROM posts ORDER BY id DESC LIMIT 1")
	if c := r.Next() ; !c {
		return nil, fmt.Errorf("could not find post in database")
	}
	*/

	/*
	verbose("updating post with new URL")
	_, err = b.db.Exec(
		"UPDATE posts SET url = ? WHERE id = ?", // this might be slow?
		p.URL, p.id,
	)
	if checkError(err) { return nil, err }
	*/

	verbose(p)
	return p, nil
}

// gets a number of posts with the characteristics described in the given
// post pointer - if l == 0; all posts are given, otherwise l posts are given
// in descending order (relative to id)
func (b *blog) getPosts(p *post, l int) ([]*post, error) {
	var s []string
	var m string

	if l != 0 {
		m = fmt.Sprintf("LIMIT %d", l)
	}

	switch {
	case p.Title != "":
		s = append(s, fmt.Sprintf("title = \"%s\"", p.Title))
	case p.Date != "":
		s = append(s, fmt.Sprintf("date_added = \"%s\"", p.Date))
	case p.id != 0:
		s = append(s, fmt.Sprint("id = ", p.id))
	}

	var e string
	if len(s) != 0 {
		e = strings.Join(s, " OR ")
		e = "WHERE " + e
	}

	u := strings.Join([]string{"SELECT * FROM posts", e, "ORDER BY id DESC", m}, " ")
	verbose(u)

	r, err := b.db.Query(u)

	posts := make([]*post, 0)
	if checkError(err) { return posts, err }
	for r.Next() {
		t := new(post)
		u, err := url.Parse(b.config.hostURL)
		if checkError(err) { return posts, err }

		err = r.Scan(&t.id, &t.Title, &t.Desc, &t.Date, &t.src)
		if checkError(err) { return posts, err }

		e, err := time.Parse(time.RFC3339, t.Date)
		if checkError(err) { return posts, err }

		t.Date = e.Format("2006-01-02")

		u.Path = path.Join(u.Path, b.config.rootFolder, "posts", fmt.Sprint(t.id))
		t.URL = u.String()

		verbose(t)
		posts = append(posts, t)
	}
	r.Close()

	return posts, nil
}

func (b *blog) updatePost(p *post) error {
	_, err := b.db.Exec("UPDATE posts SET title = ?, desc = ?, src = ? WHERE id = ?", p.Title, p.Desc, p.src, p.id)
	return err
}
