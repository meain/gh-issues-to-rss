package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestFetchRssAll(t *testing.T) {
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
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos").
		Reply(200).
		JSON(data)

	// delete any cashed file
	path := cacheLocation + "/meain/dotfiles/issues.json"
	os.Remove(path)

	request, _ := http.NewRequest(http.MethodGet, "/meain/dotfiles", nil)
	response := httptest.NewRecorder()
	handler(response, request)

	got := response.Body.String()
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

	if !strings.Contains(got, rssContent) {
		t.Fatalf("Rss feed content does not match up")
	}

}

func TestFetchRssOpenOnly(t *testing.T) {
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
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos").
		Reply(200).
		JSON(data)

	// delete any cashed file
	path := cacheLocation + "/meain/dotfiles/issues.json"
	os.Remove(path)

	request, _ := http.NewRequest(http.MethodGet, "/meain/dotfiles?m=io&m=po", nil)
	response := httptest.NewRecorder()
	handler(response, request)

	got := response.Body.String()
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

	if !strings.Contains(got, rssContent) {
		t.Fatalf("Rss feed content does not match up")
	}

}

func TestFetchRssWithGoodFirstLabel(t *testing.T) {
	data := []GithubIssue{
		GithubIssue{
			CreatedAt: "2021-09-08T12:44:47Z",
			Title:     "Sample Entry",
			HTMLURL:   "https://example.com",
			Body:      "Some body",
			Labels:    []GithubIssueLabel{GithubIssueLabel{Name: "good-first-issue"}},
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
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos").
		Reply(200).
		JSON(data)

	// delete any cashed file
	path := cacheLocation + "/meain/dotfiles/issues.json"
	os.Remove(path)

	request, _ := http.NewRequest(http.MethodGet, "/meain/dotfiles?l=good-first-issue", nil)
	response := httptest.NewRecorder()
	handler(response, request)

	got := response.Body.String()
	shouldContent := `    <item>
      <title>[issue-open]: Sample Entry</title>
      <link>https://example.com</link>
      <description>Some body</description>
      <content:encoded><![CDATA[Some body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>`

	shouldntContent := `    <item>
      <title>[issue-open]: Another Entry</title>
      <link>https://example.com</link>
      <description>Another body</description>
      <content:encoded><![CDATA[Another body]]></content:encoded>
      <pubDate>Wed, 08 Sep 2021 12:44:47 +0000</pubDate>
    </item>`

	if !strings.Contains(got, shouldContent) {
		t.Fatalf("Rss feed content does not match up")
	}
	if strings.Contains(got, shouldntContent) {
		t.Fatalf("Rss feed content unnecessary stuff")
	}

}
