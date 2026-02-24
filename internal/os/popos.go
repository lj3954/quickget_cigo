package os

import (
	"net/url"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

var popApiUrl, _ = url.Parse("https://api.pop-os.org/builds")

var PopOS = OS{
	Name:           "popos",
	PrettyName:     "Pop!_OS",
	Homepage:       "https://pop.system76.com/",
	Description:    "Operating system for STEM and creative professionals who use their computer as a tool to discover and create.",
	ConfigFunction: createPopOSConfigs,
}

func createPopOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	// Pop!_OS does not have an API that can be used to get a list of releases, so we'll just try Ubuntu's
	ubuntuReleases, err := getUbuntuReleases()
	if err != nil {
		return nil, err
	}

	ch, wg := getChannels()

	addConfig := func(release, arch string) {
		wg.Go(func() {
			baseUrl := popApiUrl.JoinPath(release)
			q := url.Values{"arch": []string{arch}}
			rawQuery := q.Encode()

			// Pop OS has switched to "Generic" over "Intel" in their API for the latest releases
			// Prefer generic if available, fallback to intel
			for _, e := range []string{"generic", "intel"} {
				url := baseUrl.JoinPath(e)
				url.RawQuery = rawQuery
				var data popApi
				// We'll ignore all errors
				if err := web.CapturePageToJson(url, &data); err != nil {
					errs <- Failure{Release: release, Arch: Arch(arch), Error: err}
					continue
				}
				if data.URL == "" {
					continue
				}
				ch <- Config{
					Release: release,
					Arch:    Arch(arch),
					ISO: []Source{
						urlChecksumSource(data.URL, data.Checksum),
					},
				}
				return
			}
		})
	}
	for _, release := range ubuntuReleases {
		for _, arch := range []string{"amd64", "arm64"} {
			addConfig(release, arch)
		}
	}
	return waitForConfigs(ch, wg), nil
}

type popApi struct {
	URL      string `json:"url"`
	Checksum string `json:"sha_sum"`
}
