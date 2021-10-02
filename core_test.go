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

func TestRssGenerationSimple(t *testing.T) {
	rssContent := `    <item>
      <title>[issue-open]: Sample Entry</title>
      <link>https://example.com</link>
      <description>Some body</description>
      <content:encoded><![CDATA[Some body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>
  </channel>
</rss>`
	data := []GithubIssue{
		GithubIssue{
			CreatedAt: "2021-09-08T12:44:47Z",
			Title:     "Sample Entry",
			HTMLURL:   "https://example.com",
			Body:      "Some body",
		},
	}
	content, err := generateRss("meain/dotfiles", data, RssModes{true, true, true, true})
	if err != nil {
		t.Fatalf("Unable to save backup file")
	}
	if !strings.Contains(content, rssContent) {
		t.Fatalf("Rss feed content does not match up")
	}
}

func TestRssGenerationWithClosed(t *testing.T) {
	rssContent := `    <item>
      <title>[issue-open]: Sample Entry</title>
      <link>https://example.com</link>
      <description>Some body</description>
      <content:encoded><![CDATA[Some body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>
    <item>
      <title>[issue-closed]: Another Entry</title>
      <link>https://example.com</link>
      <description>Another body</description>
      <content:encoded><![CDATA[Another body]]></content:encoded>
      <pubDate>Fri, 08 Oct 2021 12:44:47 +0000</pubDate>
    </item>
    <item>
      <title>[issue-open]: Another Entry</title>
      <link>https://example.com</link>
      <description>Another body</description>
      <content:encoded><![CDATA[Another body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>
  </channel>
</rss>`
	data := []GithubIssue{
		GithubIssue{
			CreatedAt: "2021-09-08T12:44:47Z",
			Title:     "Sample Entry",
			HTMLURL:   "https://example.com",
			Body:      "Some body",
		},
		GithubIssue{
			CreatedAt: "2021-09-08T12:44:47Z",
			ClosedAt:  "2021-10-08T12:44:47Z",
			State:     "closed",
			Title:     "Another Entry",
			HTMLURL:   "https://example.com",
			Body:      "Another body",
		},
	}
	content, err := generateRss("meain/dotfiles", data, RssModes{true, true, true, true})
	if err != nil {
		t.Fatalf("Unable to save backup file")
	}
	if !strings.Contains(content, rssContent) {
		t.Fatalf("Rss feed content does not match up")
	}
}

func TestRssGenerationWithClosedButOnlyOpen(t *testing.T) {
	rssContent := `    <item>
      <title>[issue-open]: Sample Entry</title>
      <link>https://example.com</link>
      <description>Some body</description>
      <content:encoded><![CDATA[Some body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>
    <item>
      <title>[issue-open]: Another Entry</title>
      <link>https://example.com</link>
      <description>Another body</description>
      <content:encoded><![CDATA[Another body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>
  </channel>
</rss>`
	data := []GithubIssue{
		GithubIssue{
			CreatedAt: "2021-09-08T12:44:47Z",
			Title:     "Sample Entry",
			HTMLURL:   "https://example.com",
			Body:      "Some body",
		},
		GithubIssue{
			CreatedAt: "2021-09-08T12:44:47Z",
			ClosedAt:  "2021-10-08T12:44:47Z",
			State:     "closed",
			Title:     "Another Entry",
			HTMLURL:   "https://example.com",
			Body:      "Another body",
		},
	}
	content, err := generateRss("meain/dotfiles", data, RssModes{true, false, true, false})
	if err != nil {
		t.Fatalf("Unable to save backup file")
	}
	if !strings.Contains(content, rssContent) {
		t.Fatalf("Rss feed content does not match up")
	}
}
