package os

import (
	"fmt"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const (
	haikuMirror    = "http://mirror.rit.edu/haiku/"
	haikuReleaseRe = `href="(r\w+)\/"`
)

type Haiku struct{}

func (Haiku) Data() OSData {
	return OSData{
		Name:        "haiku",
		PrettyName:  "Haiku",
		Homepage:    "https://www.haiku-os.org/",
		Description: "Specifically targets personal computing. Inspired by the BeOS, Haiku is fast, simple to use, easy to learn and yet very powerful.",
	}
}

func (Haiku) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(haikuMirror, haikuReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := haikuMirror + release + "/"
		iso := fmt.Sprintf("haiku-%s-x86_64-anyboot.iso", release)
		go func() {
			defer wg.Done()
			url := mirror + iso
			checksums, err := cs.Build(cs.Sha256Regex, url+".sha256")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
			checksum := checksums[iso]
			ch <- Config{
				Release: release,
				GuestOS: quickgetdata.Haiku,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}

		}()
	}
	return waitForConfigs(ch, wg), nil
}
