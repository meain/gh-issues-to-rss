package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	params := r.URL.Query()
	m, ok := params["m"]
	modes := RssModes{true, true, true, true}
	if ok {
		modes = RssModes{false, false, false, false}
		for _, entry := range m {
			switch entry {
			case "io":
				modes.IssueOpen = true
			case "ic":
				modes.IssuesClosed = true
			case "po":
				modes.PROpen = true
			case "pc":
				modes.PRClosed = true
			}
		}
	}
	splits := strings.Split(url, "/")
	if len(splits) != 3 { // url starts with /
		http.Error(w, "Invalid request: call `<url>/org/repo`", http.StatusBadRequest)
		return
	}
	repo := splits[1] + "/" + splits[2]
	rss, err := getIssueFeed(repo, modes)
	if err != nil {
		http.Error(w, "Unable to fetch atom feed", http.StatusNotFound)
		return
	}
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[OK]", repo)
	io.WriteString(w, rss)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", path.Base(os.Args[0]), "[repo] [--server]")
	} else {
		var repo = os.Args[1]
		if repo == "--help" {
			fmt.Println("Usage:", path.Base(os.Args[0]), "[repo] [--server]")
		} else if repo != "--server" {
			atom, err := getIssueFeed(repo, RssModes{true, true, true, true})
			if err != nil {
				log.Fatal("Unable to create feed for repo", repo, ":", err)
			}
			fmt.Println(atom)
		} else {
			http.HandleFunc("/", handler)

			//Use the default DefaultServeMux.
			port := os.Getenv("PORT")
			if port == "" {
				port = ":8080"
			} else {
				port = ":" + port
			}
			fmt.Println("Starting server on", port)
			err := http.ListenAndServe(port, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
