package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var baseUrl = "https://api.github.com/repos/"
var cacheLocation = "/tmp/gh-issues-to-rss-cache"

func getModesFromList(m []string) RssModes {
	modes := RssModes{false, false, false, false}
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
	return modes
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}

func handler(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method is not supported", http.StatusNotFound)
		return
	}
	url := r.URL.Path
	if url == "/" {
		path, err := os.Executable()
		if err != nil {
			log.Println(err)
		}
		data, err := ioutil.ReadFile(filepath.Dir(path) + "/index.html")
		if err != nil {
			http.Error(w, "Unable to fetch index.html", http.StatusNotFound)
			return
		}
		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(data))
		return
	}
	if url == "/_ping" {
		io.WriteString(w, "PONG")
		return
	}
	params := r.URL.Query()
	m, ok := params["m"]
	modes := RssModes{true, true, true, true}
	if ok {
		modes = getModesFromList(m)
	}

	l, ok := params["l"]
	var labels []string
	for _, label := range l {
		labels = append(labels, label)
	}
	splits := strings.Split(url, "/")
	if len(splits) != 3 { // url starts with /
		http.Error(w, "Invalid request: call `<url>/org/repo`", http.StatusBadRequest)
		return
	}
	repo := splits[1] + "/" + splits[2]
	rss, err := getIssueFeed(repo, modes, labels)
	if err != nil {
		http.Error(w, "Unable to fetch atom feed", http.StatusNotFound)
		return
	}
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[OK]", repo)
	io.WriteString(w, rss)
}

func getCliArgs() (string, RssModes, []string, bool, bool) {
	var modes string
	var labels string
	var server bool
	flag.StringVar(&modes, "m", "", "Comma separated list of modes [io,ic,po,pc]")
	flag.StringVar(&labels, "l", "", "Comma separated list of labels")
	flag.BoolVar(&server, "server", false, "display in uppercase")

	flag.Parse() // after declaring flags we need to call it

	if !server && len(flag.Args()) != 1 {
		return "", RssModes{}, nil, false, false
	}
	modeItems := RssModes{true, true, true, true}
	if modes != "" {
		modeItems = getModesFromList(strings.Split(modes, ","))
	}
	var labelItems []string
	if labels != "" {
		labelItems = strings.Split(labels, ",")
	}
	var repo = ""
	if !server {
		repo = flag.Args()[0]
	}

	return repo, modeItems, labelItems, server, true
}

func main() {
	flag.Usage = func() {
		fmt.Println(path.Base(os.Args[0]), "[FLAGS] [repo] [--server]")
		flag.PrintDefaults()
	}

	var repo, modes, labels, server, valid = getCliArgs()
	if !valid {
		flag.Usage()
		os.Exit(1)
	}
	if !server {
		atom, err := getIssueFeed(repo, modes, labels)
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
