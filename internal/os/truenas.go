package os

import (
	"regexp"
	"slices"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	truenasMirror = "https://download.sys.truenas.net/"
	truenasIsoRe  = `href="(TrueNAS-SCALE-([^-]+)\/[\d\.]+\/TrueNAS-SCALE-(\d+\.\d+)[\d\.]+\.iso)"`
)

var TrueNAS = OS{
	Name:           "truenas-community",
	PrettyName:     "TrueNAS Community Edition",
	Homepage:       "https://www.truenas.com/truenas-community-edition/",
	Description:    "TrueNAS is the world’s most popular storage OS because it’s the only universal and unified data platform, giving you the freedom to use file, block, or object storage plus container and VM support to scale according to your needs.",
	ConfigFunction: createTrueNASConfigs,
}

func createTrueNASConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(truenasMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(truenasIsoRe)
	matches := isoRe.FindAllStringSubmatch(page, -1)

	slices.Reverse(matches)
	matches = slices.CompactFunc(matches, func(a, b []string) bool {
		// Remove duplicate named releases (to filter out old patch releases)
		return a[2] == b[2]
	})
	if len(matches) > 3 {
		matches = matches[:3]
	}

	ch, wg := getChannelsWith(len(matches))
	for _, match := range matches {
		go func() {
			defer wg.Done()
			url := truenasMirror + match[1]
			release := match[3]

			// Checksums can either contain a SHA256 and nothing else, or a SHA256 and filename. We'll account for it with this manual length check (sha256 is 64 characters)
			cs, err := web.CapturePage(url + ".sha256")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
			if len(cs) > 64 {
				cs = cs[:64]
			}

			ch <- Config{
				Release: release,
				ISO: []Source{
					urlChecksumSource(url, cs),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
