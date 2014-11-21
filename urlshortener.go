package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"menteslibres.net/gosexy/redis"
	"net/http"
	"time"
)

var cli *redis.Client

var host_address string = "localhost" //

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type ShortenedURL struct {
	Code string `json: "Code"`
	URL  string `json: "URL"`
}

func (s *ShortenedURL) save() {
	cli.HSet("urls", s.Code, s.URL)
	cli.HSet("rev_urls", s.URL, s.Code)
}

// Main function
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	cli = redis.New()

	err := cli.Connect("localhost", 6379)

	if err != nil {
		log.Fatal(err)
	}

	defer cli.Save()
	defer cli.Quit()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[len("/"):] // /<here>
		if code == "" || len(code) > 5 {
			if err != nil {
				log.Fatal(err)
			}

			http.Redirect(w, r, "/create", 301)

		} else if exists, _ := cli.HExists("urls", code); exists {
			url, _ := cli.HGet("urls", code)
			http.Redirect(w, r, url, 301)
		} else {
			fmt.Fprintln(w, "Code not found!")
		}

	})

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("url.html")

		if err != nil {
			log.Fatal(err)
		}

		t.Execute(w, "/v")
	})

	http.HandleFunc("/v", func(w http.ResponseWriter, r *http.Request) {

		t, err1 := template.ParseFiles("v.html")

		if err1 != nil {
			log.Fatal(err1)
		}

		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}

		if r.FormValue("longUrl") == "" {
			fmt.Fprintln(w, "longUrl cannot be blank or empty!")
		} else {
			var url ShortenedURL
			if exists, _ := cli.HExists("rev_urls", r.FormValue("longUrl")); exists {
				code, _ := cli.HGet("rev_urls", r.FormValue("longUrl"))
				url = ShortenedURL{Code: code, URL: r.FormValue("longUrl")}
				t.Execute(w, &url)
			} else {
				url = ShortenedURL{Code: GenHash(), URL: r.FormValue("longUrl")}
				url.save()
				t.Execute(w, &url)
			}
		}
	})

	http.ListenAndServe(":80", nil)
}

func GenHash() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
