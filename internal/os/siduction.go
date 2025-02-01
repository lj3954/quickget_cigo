package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	siductionMirror   = "https://mirror.math.princeton.edu/pub/siduction/iso/"
	siductionSubdirRe = `href="([^\/\.]+)\/"`
	siductionIsoRe    = `href="(siduction-.*?\.iso)"`
)

type Siduction struct{}

func (Siduction) Data() OSData {
	return OSData{
		Name:        "siduction",
		PrettyName:  "Siduction",
		Homepage:    "https://siduction.org/",
		Description: "Operating system based on the Linux kernel and the GNU project. In addition, there are applications and libraries from Debian.",
	}
}

func (Siduction) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	subdirRe := regexp.MustCompile(siductionSubdirRe)
	release := "latest"

	subdirs, _, err := getBasicReleases(siductionMirror, subdirRe, 1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	isoRe := regexp.MustCompile(siductionIsoRe)

	for subdir := range subdirs {
		url := siductionMirror + subdir + "/"
		editions, numEditions, err := getBasicReleases(url, subdirRe, -1)
		if err != nil {
			return nil, err
		}
		wg.Add(numEditions)
		for edition := range editions {
			go func() {
				defer wg.Done()
				url := url + edition + "/"
				page, err := web.CapturePage(url)
				if err != nil {
					errs <- Failure{Release: release, Edition: edition, Error: err}
					return
				}
				isoMatch := isoRe.FindStringSubmatch(page)
				if len(isoMatch) != 2 {
					errs <- Failure{Release: release, Edition: edition, Error: fmt.Errorf("No iso found for %s", edition)}
					return
				}
				iso := isoMatch[1]
				url += iso

				checksum, err := cs.SingleWhitespace(url + ".sha256")
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
	}
	return waitForConfigs(ch, wg), nil
}
