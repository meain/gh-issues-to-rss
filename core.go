package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/feeds"
)

func makeRequest(repo string) ([]byte, error) {
	response, err := http.Get(baseUrl + repo + "/issues?state=all")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("Unable to fetch data. Is the repo valid?")
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func saveBackup(repo string, content []byte) error {
	path := cacheLocation + "/" + repo
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("Unable to create directory for caching:", err)
		}
	}

	err := ioutil.WriteFile(path+"/issues.json", content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadBackup(repo string) ([]byte, error) {
	path := cacheLocation + "/" + repo + "/issues.json"
	fi, err := os.Stat(path)
	if time.Now().Sub(fi.ModTime()) > 12*time.Hour {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRss(repo string, data []GithubIssue) (string, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:   repo,
		Link:    &feeds.Link{Href: "https://github.com/" + repo},
		Created: now,
	}

	var items []*feeds.Item
	for _, entry := range data {
		entryType := "issue"
		if entry.PullRequest.URL != "" {
			entryType = "pr"
		}
		d, err := time.Parse("2006-01-02T15:04:05Z07:00", entry.CreatedAt)
		if err != nil {
			fmt.Println("Unable to parse date:", err)
		}

		if entry.State == "closed" {
			items = append(items, &feeds.Item{
				Title:       "[" + entryType + "-" + "closed" + "]: " + entry.Title,
				Link:        &feeds.Link{Href: entry.URL},
				Description: entry.Body,
				Author:      &feeds.Author{Name: entry.User.Login},
				Created:     d,
			})
		}
		items = append(items, &feeds.Item{
			Title:       "[" + entryType + "-" + "open" + "]: " + entry.Title,
			Link:        &feeds.Link{Href: entry.URL},
			Description: entry.Body,
			Author:      &feeds.Author{Name: entry.User.Login},
			Created:     d,
		})

	}
	feed.Items = items

	rss, err := feed.ToRss()
	if err != nil {
		return "", err
	}
	return rss, nil
}

func getData(repo string) ([]byte, error) {
	content, err := loadBackup(repo)
	if err != nil || content == nil {
		fmt.Println("No cache fount for " + repo + ", fetching from Github")
		resp, err := makeRequest(repo)
		if err != nil {
			return nil, err
		}
		err = saveBackup(repo, resp)
		if err != nil {
			fmt.Println("Unable to save backup:", err)
		}
		return resp, nil
	}
	return content, nil
}

func getIssueFeed(repo string) (string, error) {
	content, err := getData(repo)
	if err != nil {
		return "", err
	}
	data := []GithubIssue{}
	if err := json.Unmarshal(content, &data); err != nil {
		return "", err
	}
	rss, err := generateRss(repo, data)
	if err != nil {
		return "", err
	}
	return rss, nil
}
