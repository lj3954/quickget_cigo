package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const endeavourMirror = "https://mirror.alpix.eu/endeavouros/iso/"

type EndeavourOS struct{}

func (EndeavourOS) Data() OSData {
	return OSData{
		Name:        "endeavouros",
		PrettyName:  "EndeavourOS",
		Homepage:    "https://endeavouros.com/",
		Description: "Provides an Arch experience without the hassle of installing it manually for both x86_64 and ARM systems.",
	}
}

func (EndeavourOS) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := capturePage(endeavourMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="(EndeavourOS_[^\d]+(\d{4}.\d{2}.\d{2}).iso)"`)
	matches := isoRe.FindAllStringSubmatch(page, -1)
	ch, wg := getChannelsWith(len(matches))
	for _, match := range matches {
		release := match[2]
		url := endeavourMirror + match[1]
		checksumUrl := url + ".sha256sum"
		go func() {
			defer wg.Done()
			checksum, err := cs.SingleWhitespace(checksumUrl)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
			ch <- Config{
				Release: release,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
