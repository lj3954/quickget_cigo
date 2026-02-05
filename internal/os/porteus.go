package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	porteusMirror = "https://mirrors.dotsrc.org/porteus/x86_64/"
	porteusIsoRe  = `^Porteus-([^-]+)-(.*?)-x86_64.iso$`
)

var Porteus = OS{
	Name:           "porteus",
	PrettyName:     "Porteus",
	Homepage:       "http://www.porteus.org/",
	Description:    "Complete linux operating system that is optimized to run from CD, USB flash drive, hard drive, or other bootable storage media.",
	ConfigFunction: createPorteusConfigs,
}

func createPorteusConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(porteusMirror)
	if err != nil {
		return nil, err
	}

	releases := make(map[string]mirror.SubDirEntry)
	for k, d := range head.SubDirs {
		if strings.HasPrefix(k, "Porteus-") {
			n := strings.TrimPrefix(k, "Porteus-")
			d.Name = n
			releases[n] = d
		}
	}

	isoRe := regexp.MustCompile(porteusIsoRe)
	ch, wg := getChannels()
	for release, d := range releases {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			checksums := make(map[string]string)
			if cf, e := contents.Files["sha256sums.txt"]; e {
				checksums, err = cs.Build(cs.Whitespace, cf)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
			}

			for f, match := range contents.FileMatches(isoRe) {
				checksum := checksums[f.Name]
				ch <- Config{
					Release: match[2],
					Edition: strings.ToLower(match[1]),
					ISO: []Source{
						webSource(f.URL.String(), checksum, "", f.Name),
					},
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
