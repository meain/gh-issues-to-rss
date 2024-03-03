package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var baseUrl = "https://api.github.com/repos/"
var cacheLocation = "/tmp/gh-issues-to-rss-cache"

//go:embed index.html
var index string

func getModesFromList(m []string) Modes {
	modes := Modes{false, false, false, false}
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

func getHandler(cacheTimeout time.Duration) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
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
			http.ServeContent(w, r, "index.html", time.Now(), strings.NewReader(index))
			return
		}
		if url == "/_ping" {
			io.WriteString(w, "PONG")
			return
		}
		params := r.URL.Query()
		m, ok := params["m"]
		modes := Modes{true, true, true, true}
		if ok {
			modes = getModesFromList(m)
		}

		splits := strings.Split(url, "/")
		if len(splits) != 3 { // url starts with /
			http.Error(w, "Invalid request: call `<url>/org/repo`", http.StatusBadRequest)
			return
		}
		repo := splits[1] + "/" + splits[2]

		labels := params["l"]
		notlabels := params["nl"]
		users := params["u"]
		notusers := params["nu"]

		rc := RunConfig{
			Modes:     modes,
			Labels:    labels,
			NotLabels: notlabels,
			Users:     users,
			NotUsers:  notusers,
			Repo:      repo,
		}

		rss, err := getIssueFeed(rc, cacheTimeout)
		if err != nil {
			http.Error(w, "Unable to fetch atom feed", http.StatusNotFound)
			return
		}
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[OK]", repo)
		io.WriteString(w, rss)
	}

	return handler
}

func getCliArgs() (config, error) {
	var (
		modes        string
		labels       string
		notlabels    string
		users        string
		notusers     string
		server       bool
		port         int
		cacheTimeout int64
	)

	flag.StringVar(&modes, "m", "", "Comma separated list of modes [io,ic,po,pc]")
	flag.StringVar(&labels, "l", "", "Comma separated list of labels to include")
	flag.StringVar(&notlabels, "nl", "", "Comma separated list of labels to exclude")
	flag.StringVar(&users, "u", "", "Comma separated list of users to include")
	flag.StringVar(&notusers, "nu", "", "Comma separated list of users to exclude")
	flag.BoolVar(&server, "server", false, "run as server instead of cli mode")
	flag.IntVar(&port, "port", 0, "port to use for server")
	flag.Int64Var(&cacheTimeout, "cache-timeout", 60*12, "cache timeout in minutes, 0 to disable")

	flag.Parse() // after declaring flags we need to call it

	if server {
		return config{ServerConfig: &ServerConfig{port, cacheTimeout}}, nil
	}

	if len(flag.Args()) != 1 {
		return config{}, errors.New("need repo when not running in server mode")
	}

	cfg := config{RunConfig: &RunConfig{
		Modes: Modes{true, true, true, true}, // default should be all true
	}}

	if modes != "" {
		cfg.RunConfig.Modes = getModesFromList(strings.Split(modes, ","))
	}

	if labels != "" { // prevents empty "" item
		cfg.RunConfig.Labels = strings.Split(labels, ",")
	}

	if notlabels != "" {
		cfg.RunConfig.NotLabels = strings.Split(notlabels, ",")
	}

	if users != "" {
		cfg.RunConfig.Users = strings.Split(users, ",")
	}

	if notusers != "" {
		cfg.RunConfig.NotUsers = strings.Split(notusers, ",")
	}

	cfg.RunConfig.Repo = flag.Args()[0]

	return cfg, nil
}

// A better version of flag.Usage
func printHelp() {
	fmt.Println(path.Base(os.Args[0]) + ` [FLAGS] [repo] [--server]

Server mode (use -server to switch to server mode):
  -port int
        port to use for server (default 8080)
  -cache-timeout float
        cache timeout in minutes, 0 to disable (default: 12 hours)
Example: ` + path.Base(os.Args[0]) + ` -server -port 8080 -cache-timeout 720

Single repo mode:
  -m string
        Comma separated list of modes [io,ic,po,pc]
  -l string
        Comma separated list of labels to include
  -nl string
        Comma separated list of labels to exclude
  -u string
        Comma separated list of users to include
  -nu string
        Comma separated list of users to exclude
Example: ` + path.Base(os.Args[0]) + ` -m io,ic,po,pc -l bug,enhancement -nl invalid -u user1,user2 -nu user3,user4 org/repo`)
}

func main() {
	flag.Usage = printHelp

	cfg, err := getCliArgs()
	if err != nil {
		flag.Usage()
		fmt.Println("\nError:", err)
		os.Exit(1)
	}

	if cfg.RunConfig != nil {
		atom, err := getIssueFeed(*cfg.RunConfig, 0)
		if err != nil {
			log.Fatal("Unable to create feed for repo", cfg.RunConfig.Repo, ":", err)
		}
		fmt.Println(atom)
	} else {
		http.HandleFunc("/", getHandler(time.Duration(cfg.ServerConfig.CacheTimeout)*time.Minute))

		port := ":" + strconv.Itoa(cfg.ServerConfig.Port)
		if cfg.ServerConfig.Port == 0 {
			port = os.Getenv("PORT")
			if port == "" {
				port = ":8080"
			} else {
				port = ":" + port
			}
		}

		fmt.Println("Starting server on", port)
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}
