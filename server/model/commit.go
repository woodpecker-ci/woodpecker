package model

type Commit struct {
	SHA      string       `json:"sha"`
	Message  string       `json:"message"`
	ForgeURL string       `json:"forge_url"`
	Author   CommitAuthor `json:"author"`
}

type CommitAuthor struct {
	Author string `json:"author"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}
