package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	antiXMirror    = "https://sourceforge.net/projects/antix-linux/files/Final/"
	antiXReleaseRe = `"name":"antiX-([0-9.]+)"`
)

var AntiX = OS{
	Name:           "antix",
	PrettyName:     "antiX",
	Homepage:       "https://antixlinux.com/",
	Description:    "Fast, lightweight and easy to install systemd-free linux live CD distribution based on Debian Stable for Intel-AMD x86 compatible systems.",
	ConfigFunction: createAntiXConfigs,
}

func createAntiXConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(antiXMirror)
	if err != nil {
		return nil, err
	}

	releases := head.ModifiedTimeSortedSubdirs()
	releases = releases[max(len(releases)-4, 0):]

	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`^antiX-[\d\.]+(?:-runit)?(?:-[^_]+)?_x64-([^.]+).iso$`)

	var addConfigs func(release string, d mirror.SubDirEntry, edition string)
	addConfigs = func(release string, d mirror.SubDirEntry, edition string) {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Edition: edition, Error: err}
				return
			}

			for k, dr := range contents.SubDirs {
				suffix := "-" + d.Name
				if prefix, ok := strings.CutSuffix(k, suffix); ok {
					addConfigs(release, dr, prefix)
				}
			}

			for f, match := range contents.FileMatches(isoRe) {
				var checksum string
				if cf, ok := contents.Files[f.Name+".sha256"]; ok {
					checksum, err = cs.SingleWhitespace(cf)
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
				}
				edition := match[1] + "-" + edition
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

	for _, d := range releases {
		release := strings.TrimPrefix(d.Name, "antiX-")
		addConfigs(release, d, "sysv")
	}

	return waitForConfigs(ch, wg), nil
}
