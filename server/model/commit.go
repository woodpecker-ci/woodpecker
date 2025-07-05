package model

type Commit struct {
	SHA      string       `json:"sha"`
	Message  string       `json:"message"`
	ForgeURL string       `json:"forge_url"`
	Author   CommitAuthor `json:"author"`
}

type CommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
