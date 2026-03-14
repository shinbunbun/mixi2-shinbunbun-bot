package github

import "time"

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Repo      Repo      `json:"repo"`
	Payload   Payload   `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	Name string `json:"name"`
}

type Payload struct {
	Size         int           `json:"size"`
	Commits      []Commit      `json:"commits"`
	Action       string        `json:"action"`
	Number       int           `json:"number"`
	PullRequest  *PullRequest  `json:"pull_request"`
	Issue        *Issue        `json:"issue"`
	RefType      string        `json:"ref_type"`
	Ref          string        `json:"ref"`
	Review       *Review       `json:"review"`
	Comment      *Comment      `json:"comment"`
	Release      *Release      `json:"release"`
	Forkee       *Forkee       `json:"forkee"`
}

type Commit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
}

type PullRequest struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
}

type Issue struct {
	Title  string `json:"title"`
	Number int    `json:"number"`
}

type Review struct {
	State string `json:"state"`
}

type Comment struct {
	Body string `json:"body"`
}

type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

type Forkee struct {
	FullName string `json:"full_name"`
}
