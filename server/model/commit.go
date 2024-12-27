package model

type Commit struct {
	SHA      string `json:"sha"`
	Message  string `json:"message"`
	ForgeURL string `json:"forge_url"`
	Author   Author `json:"author"`
}

type Author struct {
	Author string `json:"author"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}
