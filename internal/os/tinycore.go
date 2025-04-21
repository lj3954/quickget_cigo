package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	tinyCoreMirror          = "http://www.tinycorelinux.net/"
	tinyCoreDownloadPageUrl = "http://www.tinycorelinux.net/downloads.html"
	tinyCoreReleaseRe       = `href="(\d+)\.x\/x86\/(?:archive|release)\/?"`
	tinyCoreIsoRe           = `href="((\w+)-\d+\.\d+\.iso)">`
)

var TinyCore = OS{
	Name:           "tinycore",
	PrettyName:     "Tiny Core Linux",
	Homepage:       "http://www.tinycorelinux.net/",
	Description:    "Highly modular based system with community build extensions.",
	ConfigFunction: createTinyCoreConfigs,
}

func createTinyCoreConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, err := getSortedReleasesFunc(tinyCoreDownloadPageUrl, tinyCoreReleaseRe, 3, semverCompare)
	if err != nil {
		return nil, err
	}

	// We're going to have to search through both 32-bit and 64-bit x86
	ch, wg := getChannelsWith(len(releases) * 2)
	isoRe := regexp.MustCompile(tinyCoreIsoRe)

	for _, release := range releases {
		for _, arch := range []string{"x86", "x86_64"} {
			go func() {
				defer wg.Done()
				url := tinyCoreMirror + release + ".x/" + arch + "/release/"
				page, err := web.CapturePage(url)
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					return
				}
				matches := isoRe.FindAllStringSubmatch(page, -1)

				wg.Add(len(matches))
				for _, match := range matches {
					go func() {
						defer wg.Done()
						url := url + match[1]
						edition := match[2]
						checksum, err := cs.SingleWhitespace(url + ".md5.txt")
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
			}()
		}
	}
	return waitForConfigs(ch, wg), nil
}
