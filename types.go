package main

type GithubIssueLabel struct {
	Name string `json:"name"`
}

type GithubIssue struct {
	ActiveLockReason  interface{}   `json:"active_lock_reason"`
	Assignee          interface{}   `json:"assignee"`
	Assignees         []interface{} `json:"assignees"`
	AuthorAssociation string        `json:"author_association"`
	Body              string        `json:"body"`
	ClosedAt          string        `json:"closed_at"`
	Comments          int64         `json:"comments"`
	CommentsURL       string        `json:"comments_url"`
	CreatedAt         string        `json:"created_at"`
	EventsURL         string        `json:"events_url"`
	HTMLURL           string        `json:"html_url"`
	ID                int64         `json:"id"`
	Labels            []GithubIssueLabel `json:"labels"`
	LabelsURL             string      `json:"labels_url"`
	Locked                bool        `json:"locked"`
	Milestone             interface{} `json:"milestone"`
	NodeID                string      `json:"node_id"`
	Number                int64       `json:"number"`
	PerformedViaGithubApp interface{} `json:"performed_via_github_app"`
	PullRequest           struct {
		DiffURL  string `json:"diff_url"`
		HTMLURL  string `json:"html_url"`
		PatchURL string `json:"patch_url"`
		URL      string `json:"url"`
	} `json:"pull_request"`
	RepositoryURL string `json:"repository_url"`
	State         string `json:"state"`
	Title         string `json:"title"`
	UpdatedAt     string `json:"updated_at"`
	URL           string `json:"url"`
	User          struct {
		AvatarURL         string `json:"avatar_url"`
		EventsURL         string `json:"events_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		HTMLURL           string `json:"html_url"`
		ID                int64  `json:"id"`
		Login             string `json:"login"`
		NodeID            string `json:"node_id"`
		OrganizationsURL  string `json:"organizations_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		ReposURL          string `json:"repos_url"`
		SiteAdmin         bool   `json:"site_admin"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		URL               string `json:"url"`
	} `json:"user"`
}
