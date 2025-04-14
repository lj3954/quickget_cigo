package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	openIndianaMirror    = "https://dlc.openindiana.org/isos/hipster/"
	openIndianaReleaseRe = `href="./(\d{8})/"`
	openIndianaIsoRe     = `href="./(OI-hipster-([^-]+)-\d{8}.iso)"`
)

var OpenIndiana = OS{
	Name:           "openindiana",
	PrettyName:     "OpenIndiana",
	Homepage:       "https://www.openindiana.org/",
	Description:    "Community supported illumos-based operating system.",
	ConfigFunction: createOpenIndianaConfigs,
}

func createOpenIndianaConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(openIndianaMirror, openIndianaReleaseRe, 5)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(openIndianaIsoRe)
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := openIndianaMirror + release + "/"
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			matches := isoRe.FindAllStringSubmatch(page, -1)
			wg.Add(len(matches))
			for _, match := range matches {
				iso, edition := match[1], match[2]
				url := mirror + iso
				checksumUrl := url + ".sha256sum"
				go func() {
					defer wg.Done()
					checksum, err := cs.SingleWhitespace(checksumUrl)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
					ch <- Config{
						GuestOS: quickgetdata.Solaris,
						Release: release,
						Edition: edition,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				}()
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}
