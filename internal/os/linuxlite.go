package os

import (
	"errors"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

const (
	linuxliteMirror = "https://sourceforge.net/projects/linux-lite/files/"
)

var LinuxLite = OS{
	Name:           "linuxlite",
	PrettyName:     "Linux Lite",
	Homepage:       "https://www.linuxliteos.com/",
	Description:    "Your first simple, fast and free stop in the world of Linux.",
	ConfigFunction: createLinuxLiteConfigs,
}

func createLinuxLiteConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(linuxliteMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	releases := head.NameSortedSubDirs(utils.SemverCompare)
	releases = releases[max(len(releases)-5, 0):]

	addConfig := func(release string, d *mirror.Directory, f mirror.File) {
		var checksum string
		if cf, e := d.Files[f.Name+".sha256"]; e {
			checksum, err = cs.SingleWhitespace(cf)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
		}
		ch <- Config{
			Release: release,
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		}
	}

	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
			}

			for k, f := range contents.Files {
				if strings.HasSuffix(k, ".iso") {
					addConfig(release, contents, f)
					return
				}
			}

			// If only release candidate versions are available, we'll check those instead. We've already returned if there's a main release
			rcs := contents.ModifiedTimeSortedSubdirs()
			if len(rcs) == 0 {
				errs <- Failure{Release: release, Error: errors.New("no iso present in dir")}
				return
			}
			rc := rcs[len(rcs)-1]
			release += "-" + rc.Name

			contents, err = rc.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
			}

			for k, f := range contents.Files {
				if strings.HasSuffix(k, ".iso") {
					addConfig(release, contents, f)
				}
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}
