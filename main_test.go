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
	handler := getHandler(0)
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
	handler := getHandler(0)
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
	handler := getHandler(0)
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
	handler := getHandler(0)
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

// func TestCliFlagParsing(t *testing.T) {
// 	oldArgs := os.Args
// 	defer func() { os.Args = oldArgs }()

// 	tests := []struct {
// 		name      string
// 		input     []string
// 		repo      string
// 		mode      Modes
// 		labels    []string
// 		notlabels []string
// 		users     []string
// 		notusers  []string
// 		server    bool
// 	}{
// 		{"simple", []string{"meain/dotfiles"}, "meain/dotfiles", Modes{true, true, true, true}, nil, nil, nil, nil, false},
// 		{"with-labels", []string{"-l", "good-first-issue", "meain/dotfiles"}, "meain/dotfiles", Modes{true, true, true, true}, []string{"good-first-issue"}, nil, nil, nil, false},
// 		{"with-modes", []string{"-m", "ic,po", "meain/dotfiles"}, "meain/dotfiles", Modes{false, true, true, false}, nil, nil, nil, nil, false},
// 		{"with-modes-and-labels", []string{"-m", "ic,po", "-l", "good-first-issue", "meain/dotfiles"}, "meain/dotfiles", Modes{false, true, true, false}, []string{"good-first-issue"}, nil, nil, nil, false},
// 		{"with-not-labels", []string{"-nl", "good-first-issue", "meain/dotfiles"}, "meain/dotfiles", Modes{true, true, true, true}, nil, []string{"good-first-issue"}, nil, nil, false},
// 		{"with-users-and-not-users", []string{"-u", "meain", "-nu", "niaem", "meain/dotfiles"}, "meain/dotfiles", Modes{true, true, true, true}, nil, nil, []string{"meain"}, []string{"niaem"}, false},

// 	  // server
// 		{"server", []string{"--server"}, "", Modes{true, true, true, true}, nil, nil, nil, nil, true},
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			flag.CommandLine = flag.NewFlagSet("gh-issues-to-rss", flag.ExitOnError)
// 			// we need a value to set Args[0] to, cause flag begins parsing at Args[1]
// 			items := []string{"gh-issues-to-rss"}
// 			items = append(items, tc.input...)

// 			os.Args = items
// 			cfg, err := getCliArgs()
// 			if err != nil {
// 				t.Fatalf("Unable to parse cli arg")
// 			}

// 			if cfg.serverConfig

// 			if !cmp.Equal(tc.repo, cfg.runConfig.repo) {
// 				t.Fatalf("values are not the same %s", cmp.Diff(tc.repo, cfg.runConfig.repo))
// 			}
// 			if !cmp.Equal(tc.mode, cfg.runConfig.modes) {
// 				t.Fatalf("values are not the same %s", cmp.Diff(tc.mode, cfg.runConfig.modes))
// 			}
// 			if !cmp.Equal(tc.labels, cfg.runConfig.labels) {
// 				t.Fatalf("values are not the same %s", cmp.Diff(tc.labels, cfg.runConfig.labels))
// 			}
// 			if !cmp.Equal(tc.notlabels, cfg.runConfig.notLabels) {
// 				t.Fatalf("values are not the same %s", cmp.Diff(tc.notlabels, cfg.runConfig.labels))
// 			}
// 			if !cmp.Equal(tc.users, cfg.runConfig.users) {
// 				t.Fatalf("values are not the same %s", cmp.Diff(tc.users, cfg.runConfig.users))
// 			}
// 			if !cmp.Equal(tc.notusers, cfg.runConfig.notUsers) {
// 				t.Fatalf("values are not the same %s", cmp.Diff(tc.notusers, cfg.runConfig.notUsers))
// 			}
// 		})
// 	}
// }

func TestCliFlagParsing(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	table := []struct {
		name  string
		input string
		cfg   config
	}{
		{
			name:  "simple repo",
			input: "meain/dotfiles",
			cfg: config{
				RunConfig: &RunConfig{
					Repo:  "meain/dotfiles",
					Modes: Modes{true, true, true, true},
				},
			},
		},
		{
			name:  "with everything",
			input: "-m ic,po -l good-first-issue -nl bug -u meain -nu niaem meain/dotfiles",
			cfg: config{
				RunConfig: &RunConfig{
					Repo:      "meain/dotfiles",
					Modes:     Modes{false, true, true, false},
					Labels:    []string{"good-first-issue"},
					NotLabels: []string{"bug"},
					Users:     []string{"meain"},
					NotUsers:  []string{"niaem"},
				},
			},
		},
		{
			name:  "multiple items",
			input: "-m ic,po -l good-first-issue,p0 -nl bug,documentation -u meain,ain -nu niaem,nia meain/dotfiles",
			cfg: config{
				RunConfig: &RunConfig{
					Repo:      "meain/dotfiles",
					Modes:     Modes{false, true, true, false},
					Labels:    []string{"good-first-issue", "p0"},
					NotLabels: []string{"bug", "documentation"},
					Users:     []string{"meain", "ain"},
					NotUsers:  []string{"niaem", "nia"},
				},
			},
		},
		{
			name:  "server",
			input: "--server",
			cfg: config{
				ServerConfig: &ServerConfig{
					Port:         0,
					CacheTimeout: 60 * 12,
				},
			},
		},
		{
			name:  "server with args",
			input: "--server -port 8081 -cache-timeout 120",
			cfg: config{
				ServerConfig: &ServerConfig{
					Port:         8081,
					CacheTimeout: 120,
				},
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("gh-issues-to-rss", flag.ExitOnError)
			// we need a value to set Args[0] to, cause flag begins parsing at Args[1]
			items := []string{"gh-issues-to-rss"}
			input := strings.Split(tc.input, " ")
			items = append(items, input...)

			os.Args = items
			cfg, err := getCliArgs()
			if err != nil {
				t.Fatalf("unable to parse cli arg: %s", err)
			}

			if !cmp.Equal(tc.cfg, cfg) {
				t.Fatalf("values are not the same %s", cmp.Diff(tc.cfg, cfg))
			}
		})
	}
}
