package main

type GithubInstallationRepositoriesMessage struct {
	Action string `json:"action"`
	Sender Sender `json:"sender"`

	RepositoriesRemoved []Repository `json:"repositories_removed"`
	RepositoriesAdded   []Repository `json:"repositories_added"`
}

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

type Sender struct {
	Login string `json:"login"`
}

type RepoUpdatedMessage struct {
	RepoName string `json:"repo_name"`
	User     string `json:"user"`
}
