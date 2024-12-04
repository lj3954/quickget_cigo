package data

type GithubAPI struct {
	TagName    string        `json:"tag_name"`
	Assets     []GithubAsset `json:"assets"`
	Prerelease bool          `json:"prerelease"`
	Body       string        `json:"body"`
}

type GithubAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}
