package os

import (
	"errors"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

const (
	archcraftMirror = "https://sourceforge.net/projects/archcraft/files/"
)

var Archcraft = OS{
	Name:           "archcraft",
	PrettyName:     "Archcraft",
	Homepage:       "https://archcraft.io/",
	Description:    "Yet another minimal Linux distribution, based on Arch Linux.",
	ConfigFunction: createArchcraftConfigs,
}

func createArchcraftConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(archcraftMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	releases := head.NameSortedSubDirs(utils.SemverCompare)
	releases = releases[max(len(releases)-3, 0):]

	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			f, e := contents.FindFile(func(f mirror.File) bool {
				return strings.HasSuffix(f.Name, ".iso")
			})
			if !e {
				errs <- Failure{Release: release, Error: errors.New("could not find ISO in directory")}
				return
			}

			var checksum string
			cf, e := contents.FindFile(func(f2 mirror.File) bool {
				return strings.HasPrefix(f2.Name, f.Name) && strings.HasSuffix(f2.Name, "sum")
			})
			if e {
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
		})
	}

	return waitForConfigs(ch, wg), nil
}
