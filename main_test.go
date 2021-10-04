package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/h2non/gock.v1"
)

func TestWebserverPing(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/_ping", nil)
	response := httptest.NewRecorder()
	handler(response, request)
	got := response.Body.String()

	if !strings.Contains(got, "PONG") {
		t.Fatalf("Did not get PONG back from server")
	}

}

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

func TestCliFlagParsing(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name   string
		input  []string
		repo   string
		mode   RssModes
		labels []string
		server bool
	}{
		{"simple", []string{"meain/dotfiles"}, "meain/dotfiles", RssModes{true, true, true, true}, nil, false},
		{"with-labels", []string{"-l", "good-first-issue", "meain/dotfiles"}, "meain/dotfiles", RssModes{true, true, true, true}, []string{"good-first-issue"}, false},
		{"with-modes", []string{"-m", "ic,po", "meain/dotfiles"}, "meain/dotfiles", RssModes{false, true, true, false}, nil, false},
		{"with-modes-and-labels", []string{"-m", "ic,po", "-l", "good-first-issue", "meain/dotfiles"}, "meain/dotfiles", RssModes{false, true, true, false}, []string{"good-first-issue"}, false},
		{"server", []string{"--server"}, "", RssModes{true, true, true, true}, nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("gh-issues-to-rss", flag.ExitOnError)
			// we need a value to set Args[0] to, cause flag begins parsing at Args[1]
			items := []string{"gh-issues-to-rss"}
			for _, i := range tc.input {
				items = append(items, i)
			}
			os.Args = items
			var repo, mode, labels, server, valid = getCliArgs()
			if !valid {
				t.Fatalf("Unable to parse cli arg")
			}
			if !cmp.Equal(tc.repo, repo) {
				t.Fatalf("values are not the same %s", cmp.Diff(tc.repo, repo))
			}
			if !cmp.Equal(tc.mode, mode) {
				t.Fatalf("values are not the same %s", cmp.Diff(tc.mode, mode))
			}
			if !cmp.Equal(tc.labels, labels) {
				t.Fatalf("values are not the same %s", cmp.Diff(tc.labels, labels))
			}
			if !cmp.Equal(tc.server, server) {
				t.Fatalf("values are not the same %s", cmp.Diff(tc.server, server))
			}
		})
	}
}
