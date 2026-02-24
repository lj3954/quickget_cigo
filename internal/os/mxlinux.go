package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	mxlinuxMirror = "https://sourceforge.net/projects/mx-linux/files/Final/"
)

var MXLinux = OS{
	Name:           "mxlinux",
	PrettyName:     "MX Linux",
	Homepage:       "https://mxlinux.org/",
	Description:    "Designed to combine elegant and efficient desktops with high stability and solid performance.",
	ConfigFunction: createMXLinuxConfigs,
}

func createMXLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(mxlinuxMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`^MX-([\d\.]+)(_\w+)?_x64.iso$`)

	for edition, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Edition: edition, Error: err}
				return
			}
			for f, match := range contents.FileMatches(isoRe) {
				release := match[1]

				var checksum string
				if cf, ok := contents.Files[f.Name+".sha256"]; ok {
					checksum, err = cs.SingleWhitespace(cf)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
				}

				if len(match[2]) > 1 {
					edition = match[2][1:]
				}

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
