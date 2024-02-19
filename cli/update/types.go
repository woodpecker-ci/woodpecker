package update

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type NewVersion struct {
	Version  string
	AssetURL string
}

const githubReleaseURL = "https://api.github.com/repos/woodpecker-ci/woodpecker/releases/latest"
