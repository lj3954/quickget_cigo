package os

import (
	"errors"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	gnomeosMirror = "https://download.gnome.org/gnomeos/"
)

var GnomeOS = OS{
	Name:           "gnomeos",
	PrettyName:     "GNOME OS",
	Homepage:       "https://os.gnome.org/",
	Description:    "Alpha nightly bleeding edge distro of GNOME",
	ConfigFunction: createGnomeOSConfigs,
}

func createGnomeOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(gnomeosMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	releases := head.NameSortedSubDirs(strings.Compare)
	releases = releases[max(len(releases)-6, 0):]

	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			f, ok := contents.FindFile(func(f mirror.File) bool {
				return strings.HasSuffix(f.Name, ".iso")
			})
			if !ok {
				errs <- Failure{Release: release, Error: errors.New("no ISO found")}
				return
			}
			ch <- Config{
				Release: release,
				ISO: []Source{
					webSource(f.URL.String(), "", "", f.Name),
				},
			}
		})
	}

	configs := waitForConfigs(ch, wg)
	configs = append(configs, Config{
		Release: "nightly",
		ISO: []Source{
			urlSource("https://os.gnome.org/download/latest/gnome_os_installer.iso"),
		},
	})
	return configs, nil
}
