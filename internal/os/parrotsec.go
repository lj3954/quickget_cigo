package os

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	parrotSecMirror    = "https://download.parrot.sh/parrot/iso/"
	parrotSecReleaseRe = `href="(\d+\.\d+(?:\.\d+)?)/"`
	parrotSecIsoRe     = `href="(Parrot-([^-]+)-[\d\.]+_([^\.]+)\.(iso|qcow2.xz))"`
)

var parrotSec = OS{
	Name:           "parrotsec",
	PrettyName:     "Parrot Security",
	Homepage:       "https://www.parrotsec.org/",
	Description:    `Provides a huge arsenal of tools, utilities and libraries that IT and security professionals can use to test and assess the security of their assets in a reliable, compliant and reproducible way.`,
	ConfigFunction: createParrotSecConfigs,
}

func createParrotSecConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, err := getSortedReleasesFunc(parrotSecMirror, parrotSecReleaseRe, 3, semverCompare)
	if err != nil {
		return nil, err
	}

	ch, wg := getChannels()
	isoRe := regexp.MustCompile(parrotSecIsoRe)

	for _, release := range slices.Backward(releases) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			url := parrotSecMirror + release + "/"
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
					config := Config{
						Release: release,
						Edition: match[2],
						Arch:    Arch(match[3]),
					}

					qcowXz := match[4] == "qcow2.xz"
					if qcowXz {
						config.Edition += "-preinstalled"
					}

					var checksum string
					checksumPage, err := web.CapturePage(url + ".hashes")
					if err != nil {
						csErrs <- Failure{Release: release, Edition: config.Edition, Error: err}
					} else if len(checksumPage) < 5 {
						csErrs <- Failure{Release: release, Error: errors.New("Unexpected checksum page length")}
					} else {
						checksum, err = cs.BuildSingleWhitespace(strings.Split(checksumPage, "\n")[4])
						if err != nil {
							csErrs <- Failure{Release: release, Edition: config.Edition, Error: err}
						}
					}

					if qcowXz {
						config.DiskImages = []Disk{
							{
								Source: webSource(url, checksum, quickgetdata.Xz, ""),
							},
						}
					} else {
						config.ISO = []Source{
							urlChecksumSource(url, checksum),
						}
					}

					ch <- config
				}()
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
