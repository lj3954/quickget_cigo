package os

import (
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const popApiUrl = "https://api.pop-os.org/builds/"

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

	addConfig := func(release, edition string, popEdition string) {
		wg.Go(func() {
			url := popApiUrl + release + "/" + popEdition
			var data popApi
			if err := web.CapturePageToJson(url, &data); err != nil {
				errs <- Failure{Release: release, Edition: edition, Error: err}
				return
			}
			// We'll ignore empty entries without logging an error; most of Ubuntu's releases won't be available
			// The error above is logged since it will only occur if the API is down or if the JSON is malformed
			if data.URL == "" {
				return
			}
			ch <- Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					urlChecksumSource(data.URL, data.Checksum),
				},
			}
		})
	}
	for _, release := range ubuntuReleases {
		addConfig(release, "standard", "intel")
		addConfig(release, "nvidia", "nvidia")
	}
	return waitForConfigs(ch, wg), nil
}

type popApi struct {
	URL      string `json:"url"`
	Checksum string `json:"sha_sum"`
}
