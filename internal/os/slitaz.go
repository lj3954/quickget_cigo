package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	slitazMirror = "https://mirror.slitaz.org/iso/rolling/"
	slitazIsoRe  = `href='(slitaz-(rolling(?:-.*?)?)\.iso)'`
)

var Slitaz = OS{
	Name:           "slitaz",
	PrettyName:     "SliTaz",
	Homepage:       "https://www.slitaz.org/",
	Description:    "Simple, fast and low resource Linux OS for servers & desktops.",
	ConfigFunction: createSlitazConfigs,
}

func createSlitazConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(slitazMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(slitazIsoRe)
	matches := isoRe.FindAllStringSubmatch(page, -1)

	ch, wg := getChannelsWith(len(matches))
	release := "latest"
	for _, match := range matches {
		go func() {
			defer wg.Done()
			edition := match[2]
			url := slitazMirror + match[1]

			// Ensure we don't panic on too short URLs
			if len(url) < 4 {
				return
			}
			checksum, err := cs.SingleWhitespace(url[:len(url)-4] + ".md5")
			if err != nil {
				csErrs <- Failure{Release: release, Edition: edition, Error: err}
			}

			ch <- Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
