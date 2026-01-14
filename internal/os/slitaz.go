package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	slitazMirror = "https://mirror.slitaz.org/iso/rolling/"
	slitazIsoRe  = `^(slitaz-(rolling(?:-.*?)?))\.iso$`
)

var Slitaz = OS{
	Name:           "slitaz",
	PrettyName:     "SliTaz",
	Homepage:       "https://www.slitaz.org/",
	Description:    "Simple, fast and low resource Linux OS for servers & desktops.",
	ConfigFunction: createSlitazConfigs,
}

func createSlitazConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(slitazMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(slitazIsoRe)

	ch, wg := getChannels()
	release := "latest"
	for f, match := range head.FileMatches(isoRe) {
		wg.Go(func() {
			edition := match[2]

			var checksum string
			if f, e := head.Files[match[1]+".md5"]; e {
				checksum, err = cs.SingleWhitespace(f.URL)
				if err != nil {
					csErrs <- Failure{Release: release, Edition: edition, Error: err}
				}
			}

			ch <- Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					webSource(f.URL, checksum, "", f.Name),
				},
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
