package main

type RepoAddedMessage struct {
	RepoName string `json:"repo_name"`
	User     string `json:"user"`
}
