package os

import (
	"iter"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	spiralLinuxMirror    = "https://sourceforge.net/projects/spirallinux/files/"
	spiralLinuxReleaseRe = `"name">(\d+\.\d+)<\/span>`
	spiralLinuxIsoRe     = `"name":"(SpiralLinux_(\w+)_\d+\.\d+_x86-64.iso)".*?sha1":"([0-9a-f]{40})"`
)

var SpiralLinux = OS{
	Name:           "spirallinux",
	PrettyName:     "SpiralLinux",
	Homepage:       "https://spirallinux.github.io/",
	Description:    "Selection of Linux spins built from Debian GNU/Linux, with a focus on simplicity and out-of-the-box usability across all the major desktop environments.",
	ConfigFunction: createSpiralLinuxConfigs,
}

func createSpiralLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(spiralLinuxMirror, spiralLinuxReleaseRe, 3)
	if err != nil {
		return nil, err
	}

	ch, wg := getChannelsWith(numReleases)
	isoRe := regexp.MustCompile(spiralLinuxIsoRe)
	for release := range releases {
		go func() {
			defer wg.Done()
			configs, err := getSpiralLinuxConfigs(release, isoRe)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			for c := range configs {
				ch <- c
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}

func getSpiralLinuxConfigs(release string, isoRe *regexp.Regexp) (iter.Seq[Config], error) {
	url := spiralLinuxMirror + release + "/"
	page, err := web.CapturePage(url)
	if err != nil {
		return nil, err
	}
	return func(yield func(Config) bool) {
		for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
			url := url + match[1] + "/download"
			edition := match[2]
			checksum := match[3]

			c := Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}

			if !yield(c) {
				break
			}
		}
	}, nil
}
