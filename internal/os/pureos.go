package os

import (
	"errors"
	"strconv"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	pureOsMirror    = "https://downloads.puri.sm/"
	pureOsReleaseRe = `href="(\d+)/"`
	pureOsEditionRe = `href="(\w.*?)/"`
	pureOsDateRe    = `href="(\d{4}-\d{2}-\d{2})/"`
	pureOsIsoRe     = `href="(pureos-([\d\.]+)-.*?-\d{8}_amd64.iso)"`
)

var PureOS = OS{
	Name:           "pureos",
	PrettyName:     "PureOS",
	Homepage:       "https://www.pureos.net/",
	Description:    "PureOS is a fully free/libre and open source GNU/Linux operating system, endorsed by the Free Software Foundation.",
	ConfigFunction: createPureOSConfigs,
}

func createPureOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(pureOsMirror)
	if err != nil {
		return nil, err
	}

	// Remove non-integer releases. They may cause duplicates and are unnecessary
	for release := range head.SubDirs {
		if _, err := strconv.Atoi(release); err != nil {
			delete(head.SubDirs, release)
		}
	}

	ch, wg := getChannels()

	for release, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
			}

			for edition, d := range contents.SubDirs {
				contents, err := d.Fetch()
				if err != nil {
					errs <- Failure{Release: release, Edition: edition, Error: err}
					return
				}

				dates := contents.ModifiedTimeSortedSubdirs()
				if len(dates) > 0 {
					contents, err = dates[len(dates)-1].Fetch()
					if err != nil {
						errs <- Failure{Release: release, Edition: edition, Error: err}
						return
					}
				}

				f, e := contents.FindFile(func(f mirror.File) bool {
					return strings.HasSuffix(f.Name, ".iso")
				})
				if !e {
					errs <- Failure{Release: release, Edition: edition, Error: errors.New("could not find ISO in mirror")}
					return
				}

				checksums := make(map[string]string)
				cf, e := contents.FindFile(func(f2 mirror.File) bool {
					isoName := strings.TrimSuffix(f.Name, ".iso")
					return strings.Contains(f2.Name, isoName) && strings.Contains(f2.Name, "sha256")
				})
				if e {
					checksums, err = cs.Build(cs.Whitespace, cf)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
				}

				checksum := checksums[f.Name]

				ch <- Config{
					Release: release,
					Edition: edition,
					ISO: []Source{
						webSource(f.URL.String(), checksum, "", f.Name),
					},
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
