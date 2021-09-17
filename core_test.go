package main

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

// Not sure if testing this has a point
func TestMakeRequest(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/issues").
		Reply(200).BodyString("mango")

	content, err := makeRequest("meain/dotfiles")
	if err != nil {
		t.Fatalf("Unable to fetch star count")
	}
	if string(content) != "mango" {
		t.Fatalf("makeReqest seems to have some problems. expected %v, got %v", "mango", string(content))
	}
}

func TestBackup(t *testing.T) {
	cacheLocationBackup := cacheLocation
	defer func() { cacheLocation = cacheLocationBackup }()
	dir, err := ioutil.TempDir("", "gh-issues-to-rss")
	if err != nil {
		log.Fatal("Unable to create temp directory:", err)
	}
	cacheLocation = dir

	err = saveBackup("meain/dotfiles", []byte("dummy"))
	if err != nil {
		t.Fatalf("Unable to save backup file")
	}
	content, err := loadBackup("meain/dotfiles")
	if err != nil {
		t.Fatalf("Unable to load backup file")
	}
	if string(content) != "dummy" {
		t.Fatalf("Invalid backup content")
	}
}

func TestRssGeneration(t *testing.T) {
	rssContent := `    <item>
      <title>[issue]: Sample Entry</title>
      <link>https://example.com</link>
      <description>Some body</description>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>
  </channel>
</rss>`
	data := []GithubIssue{
		GithubIssue{
			UpdatedAt: "2021-09-08T12:44:47Z",
			Title:     "Sample Entry",
			URL:       "https://example.com",
			Body:      "Some body",
		},
	}
	content, err := generateRss("meain/dotfiles", data)
	if err != nil {
		t.Fatalf("Unable to save backup file")
	}
	if !strings.Contains(content, rssContent) {
		t.Fatalf("Rss feed content does not match up")
	}
}
