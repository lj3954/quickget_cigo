package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	arcoLinuxMirror = "https://sourceforge.net/projects/arconetpro/files/"
	arcoLinuxIsoRe  = `^(?:arco|arch)[^-]+-(?:20|v)(\d{2}\.\d{2}\.\d{2})-x86_64\.iso$`
)

var ArcoLinux = OS{
	Name:           "arcolinux",
	PrettyName:     "ArcoLinux",
	Homepage:       "https://arcolinux.com/",
	Description:    "It's all about becoming an expert in Linux.",
	ConfigFunction: createArcoLinuxConfigs,
}

func createArcoLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(arcoLinuxMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(arcoLinuxIsoRe)
	ch, wg := getChannels()

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
				if cf, e := contents.Files[f.Name+".md5"]; e {
					checksum, err = cs.SingleWhitespace(cf)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
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
