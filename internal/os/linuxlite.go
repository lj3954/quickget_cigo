package os

import (
	"fmt"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	linuxliteMirror    = "https://sourceforge.net/projects/linux-lite/files/"
	linuxliteReleaseRe = `"name":"(\d(?:\.\d+)+)"`
)

type LinuxLite struct{}

func (LinuxLite) Data() OSData {
	return OSData{
		Name:        "linuxlite",
		PrettyName:  "Linux Lite",
		Homepage:    "https://www.linuxliteos.com/",
		Description: "Your first simple, fast and free stop in the world of Linux.",
	}
}

func (LinuxLite) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getSortedReleases(linuxliteMirror, linuxliteReleaseRe, 5)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(len(releases))

	for _, release := range releases {
		urlBase := fmt.Sprintf("%s%s/linux-lite-%s-64bit.iso", linuxliteMirror, release, release)
		url := urlBase + "/download"
		checksumUrl := urlBase + ".sha256/download"
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
