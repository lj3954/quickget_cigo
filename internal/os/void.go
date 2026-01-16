package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	voidMirror = "https://repo-default.voidlinux.org/live/"
	voidIsoRe  = `^void-live-(aarch64|x86_64)(-musl)?-\d{8}-(.*?)\.iso$`
)

var Void = OS{
	Name:           "void",
	PrettyName:     "Void Linux",
	Homepage:       "https://voidlinux.org/",
	Description:    "General purpose operating system. Its package system allows you to quickly install, update and remove software; software is provided in binary packages or can be built directly from sources.",
	ConfigFunction: createVoidConfigs,
}

func createVoidConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(voidMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(voidIsoRe)
	ch, wg := getChannels()

	// Current overlaps with a named release. Remove it ahead of time
	delete(head.SubDirs, "current")
	releases := head.NameSortedSubDirs(strings.Compare)
	releases = releases[max(len(releases)-3, 0):]

	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			contents, err := d.Fetch(c)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			checksums := make(map[string]string)
			if f, e := contents.Files["sha256sum.txt"]; e {
				checksums, err = cs.Build(cs.Sha256Regex, f.URL)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
			}

			for f, match := range contents.FileMatches(isoRe) {
				edition := match[3] + match[2]
				checksum := checksums[f.Name]
				ch <- Config{
					Release: release,
					Edition: edition,
					Arch:    Arch(match[1]),
					ISO: []Source{
						webSource(f.URL, checksum, "", f.Name),
					},
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
