package main

type RepoUpdatedMessage struct {
	RepoName string `json:"repo_name"`
	User     string `json:"user"`
}
