package os

import (
	"errors"
	"regexp"
)

const (
	gnomeosMirror    = "https://download.gnome.org/gnomeos/"
	gnomeosReleaseRe = `href="(\d[^/]+)\/"`
)

type GnomeOS struct{}

func (GnomeOS) Data() OSData {
	return OSData{
		Name:        "gnomeos",
		PrettyName:  "GNOME OS",
		Homepage:    "https://os.gnome.org/",
		Description: "Alpha nightly bleeding edge distro of GNOME",
	}
}

func (GnomeOS) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(gnomeosMirror, gnomeosReleaseRe, 6)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases)
	isoRe := regexp.MustCompile(`href="(gnome_os.*?.iso)"`)

	for release := range releases {
		mirror := gnomeosMirror + release + "/"
		go func() {
			defer wg.Done()
			page, err := capturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			isoMatch := isoRe.FindStringSubmatch(page)
			if isoMatch == nil {
				errs <- Failure{Release: release, Error: errors.New("No ISO found")}
				return
			}
			url := mirror + isoMatch[1]
			ch <- Config{
				Release: release,
				ISO: []Source{
					urlSource(url),
				},
			}
		}()
	}

	configs := waitForConfigs(ch, wg)
	configs = append(configs, Config{
		Release: "nightly",
		ISO: []Source{
			urlSource("https://os.gnome.org/download/latest/gnome_os_installer.iso"),
		},
	})
	return configs, nil
}
