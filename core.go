package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"io"

	"io/fs"

	"github.com/gorilla/feeds"
)

func makeRequest(repo string) ([]byte, error) {
	// do an http get request to the github api. Add auth header if token is present
	req, err := http.NewRequest("GET", baseUrl+repo+"/issues?state=all", nil)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("GH_ISSUES_TO_RSS_GITHUB_TOKEN")
	if token != "" {
		fmt.Println("adding auth token")
		req.Header.Add("Authorization", "Bearer"+token)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New("unable to fetch data, make sure you have a valid repo")
	}

	defer response.Body.Close()

	body, err := io.ReadAll(io.Reader(response.Body))
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

	err := os.WriteFile(path+"/issues.json", content, fs.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

func loadBackup(repo string, timeout time.Duration) ([]byte, error) {
	path := cacheLocation + "/" + repo + "/issues.json"

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if time.Now().Sub(fi.ModTime()) > timeout {
		return nil, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func isIn(item string, items []string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}

func generateRss(data []GithubIssue, rc RunConfig) (string, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:   rc.Repo,
		Link:    &feeds.Link{Href: "https://github.com/" + rc.Repo},
		Created: now,
	}

	var items []*feeds.Item
	fdata := filterIssues(data, rc)

	for _, entry := range fdata {
		entryType := "issue"
		if entry.PullRequest.URL != "" {
			entryType = "pr"
		}
		createTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", entry.CreatedAt)
		closeTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", entry.ClosedAt)

		if entry.State == "closed" {
			if !((entryType == "pr" && !rc.Modes.PRClosed) || (entryType == "issue" && !rc.Modes.IssuesClosed)) {
				items = append(items, &feeds.Item{
					Title:       "[" + entryType + "-" + "closed" + "]: " + entry.Title,
					Link:        &feeds.Link{Href: entry.HTMLURL},
					Description: strings.ReplaceAll(entry.Body, "\n", "<br>"),
					Content:     strings.ReplaceAll(entry.Body, "\n", "<br>"),
					Author:      &feeds.Author{Name: entry.User.Login},
					Created:     closeTime,
				})
			}
		}
		if (entryType == "pr" && !rc.Modes.PROpen) || (entryType == "issue" && !rc.Modes.IssueOpen) {
			continue
		}
		items = append(items, &feeds.Item{
			Title:       "[" + entryType + "-" + "open" + "]: " + entry.Title,
			Link:        &feeds.Link{Href: entry.HTMLURL},
			Description: strings.ReplaceAll(entry.Body, "\n", "<br>"),
			Content:     strings.ReplaceAll(entry.Body, "\n", "<br>"),
			Author:      &feeds.Author{Name: entry.User.Login},
			Created:     createTime,
		})

	}
	feed.Items = items

	rss, err := feed.ToRss()
	if err != nil {
		return "", err
	}
	return rss, nil
}

func getData(repo string, cacheTimeout time.Duration) ([]byte, error) {
	content, err := loadBackup(repo, cacheTimeout)
	if err != nil || content == nil {
		fmt.Println("No cache found for " + repo + ", fetching from Github")
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

func getIssueFeed(rc RunConfig, cacheTimeout time.Duration) (string, error) {
	content, err := getData(rc.Repo, cacheTimeout)
	if err != nil {
		return "", err
	}

	data := []GithubIssue{}
	if err := json.Unmarshal(content, &data); err != nil {
		return "", err
	}

	rss, err := generateRss(data, rc)
	if err != nil {
		return "", err
	}
	return rss, nil
}

func filterIssues(issues []GithubIssue, rc RunConfig) []GithubIssue {
	var fi []GithubIssue

	for _, issue := range issues {
		var issueLabels []string
		for _, i := range issue.Labels {
			issueLabels = append(issueLabels, i.Name)
		}

		bk := false
		for _, label := range rc.NotLabels {
			if isIn(label, issueLabels) {
				bk = true
				break
			}
		}

		if bk {
			continue
		}

		for _, user := range rc.NotUsers {
			if issue.User.Login == user {
				bk = true
				break
			}
		}

		if bk {
			continue
		}

		for _, label := range rc.Labels {
			if !isIn(label, issueLabels) {
				bk = true
				break
			}
		}

		if bk {
			continue
		}

		bk = len(rc.Users) != 0

		for _, user := range rc.Users {
			if issue.User.Login == user {
				bk = false
				break
			}
		}

		if bk {
			continue
		}

		fi = append(fi, issue)
	}

	return fi
}
