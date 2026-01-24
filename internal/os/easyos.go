package os

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	easyosMirror = "https://distro.ibiblio.org/easyos/amd64/releases/"
)

var EasyOS = OS{
	Name:           "easyos",
	PrettyName:     "EasyOS",
	Homepage:       "https://easyos.org/",
	Description:    "Experimental distribution designed from scratch to support containers.",
	ConfigFunction: createEasyOSConfigs,
}

func createEasyOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	releases, err := getEasyOSReleases(errs, c)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(releases, func(a, b mirror.SubDirEntry) int {
		return utils.SemverCompare(a.Name, b.Name)
	})
	releases = releases[max(len(releases)-3, 0):]

	ch, wg := getChannels()
	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			var checksum string
			if f, e := contents.Files["md5sum.txt"]; e {
				checksum, err = cs.SingleWhitespace(f)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
			}

			for k, f := range contents.Files {
				if strings.HasSuffix(k, ".img") || strings.HasSuffix(k, ".img.gz") {
					var archiveFormat quickgetdata.ArchiveFormat
					if strings.HasSuffix(f.Name, ".img.gz") {
						archiveFormat = quickgetdata.Gz
					}
					ch <- Config{
						Release: release,
						DiskImages: []Disk{
							{
								Source: webSource(f.URL.String(), checksum, archiveFormat, f.Name),
								Format: quickgetdata.Raw,
							},
						},
					}
					return
				}
			}

		})
	}

	return waitForConfigs(ch, wg), nil
}

func getEasyOSReleases(errs chan<- Failure, c mirror.Client) ([]mirror.SubDirEntry, error) {
	contents, err := c.ReadDir(easyosMirror)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	ch := make(chan mirror.SubDirEntry)
	var releases []mirror.SubDirEntry
	go func() {
		wg.Wait()
		close(ch)
	}()

	for _, d := range contents.SubDirs {
		wg.Go(func() {
			yearsDir, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: d.Name, Error: err}
				return
			}

			years := slices.Collect(maps.Values(yearsDir.SubDirs))
			if len(years) == 0 {
				errs <- Failure{Release: d.Name, Error: errors.New("no years found in directory")}
			}
			latestYear := slices.MaxFunc(years, func(a, b mirror.SubDirEntry) int {
				return strings.Compare(a.Name, b.Name)
			})

			releasesDir, err := latestYear.Fetch()
			if err != nil {
				errs <- Failure{Release: d.Name, Error: err}
				return
			}

			releases := slices.Collect(maps.Values(releasesDir.SubDirs))
			if len(releases) == 0 {
				errs <- Failure{Release: d.Name, Error: errors.New("no releases found in year directory")}
			}
			latestRelease := slices.MaxFunc(releases, func(a, b mirror.SubDirEntry) int {
				return utils.SemverCompare(a.Name, b.Name)
			})
			ch <- latestRelease
		})
	}

	for release := range ch {
		releases = append(releases, release)
	}

	return releases, nil
}
