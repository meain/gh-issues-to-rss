package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var baseUrl = "https://api.github.com/repos/"
var cacheLocation = "/tmp/gh-issues-to-rss-cache"

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported", http.StatusNotFound)
		return
	}
	url := r.URL.Path
	splits := strings.Split(url, "/")
	if len(splits) != 3 { // url starts with /
		http.Error(w, "Invalid request: call `<url>/org/repo`", http.StatusBadRequest)
		return
	}
	repo := splits[1] + "/" + splits[2]
	rss, err := getIssueFeed(repo)
	if err != nil {
		http.Error(w, "Unable to fetch atom feed", http.StatusNotFound)
		return
	}
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[OK]", repo)
	io.WriteString(w, rss)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "[repo] [--server]")
	} else {
		var repo = os.Args[1]
		if repo != "--server" {
			atom, err := getIssueFeed(repo)
			if err != nil {
				log.Fatal("Unable to create feed for repo", repo, ":", err)
			}
			fmt.Println(atom)
		} else {
			http.HandleFunc("/", handler)

			//Use the default DefaultServeMux.
			fmt.Println("Starting server on :8080")
			err := http.ListenAndServe(":8080", nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
