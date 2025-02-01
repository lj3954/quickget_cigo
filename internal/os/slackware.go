package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	slackwareMirror    = "https://slackware.nl/slackware/slackware-iso/"
	slackwareReleaseRe = `href="(slackware64-(\d+\.\d+)-iso)/"`
)

type Slackware struct{}

func (Slackware) Data() OSData {
	return OSData{
		Name:        "slackware",
		PrettyName:  "Slackware",
		Homepage:    "https://www.slackware.com/",
		Description: "Advanced Linux operating system, designed with the twin goals of ease of use and stability as top priorities.",
	}
}

func (Slackware) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(slackwareMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(slackwareReleaseRe)
	matches := releaseRe.FindAllStringSubmatch(page, -1)

	ch, wg := getChannelsWith(len(matches))

	for _, match := range matches {
		go func() {
			defer wg.Done()
			url := slackwareMirror + match[1] + "/"
			release := match[2]
			iso := fmt.Sprintf("slackware64-%s-install-dvd.iso", release)
			url += iso
			checksum, err := cs.SingleWhitespace(url + ".md5")
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
