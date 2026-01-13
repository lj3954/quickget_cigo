package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	devuanMirror = "https://files.devuan.org/"
)

var Devuan = OS{
	Name:           "devuan",
	PrettyName:     "Devuan",
	Homepage:       "https://devuan.org/",
	Description:    "Fork of Debian without systemd that allows users to reclaim control over their system by avoiding unnecessary entanglements and ensuring Init Freedom.",
	ConfigFunction: createDevuanConfigs,
}

func createDevuanConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(devuanMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	isoRe := regexp.MustCompile(`^devuan_[a-zA-Z]+_([0-9.]+)_amd64_desktop-live.iso$`)

	for k, releaseDir := range head.SubDirs {
		if !strings.HasPrefix(k, "devuan_") {
			continue
		}
		release := strings.TrimPrefix(releaseDir.Name, "devuan_")
		wg.Go(func() {
			contents, err := releaseDir.Fetch(c)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			// If there's a desktop live subdirectory we'll use it, as that's the standard directory
			// structure as of now. Otherwise, just try with the main directory
			if d, e := contents.SubDirs["desktop-live"]; e {
				contents, err = d.Fetch(c)
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					return
				}
			}

			checksums := make(map[string]string)
			for k, f := range contents.Files {
				k = strings.ToLower(k)
				if strings.HasSuffix(k, "txt") && strings.Contains(k, "sum") {
					checksums, err = cs.Build(cs.Whitespace, f.URL)
					if err != nil {
						csErrs <- Failure{Release: release, Error: err}
					} else {
						break
					}
				}
			}

			for f := range contents.MatchingFiles(isoRe) {
				checksum := checksums[f.Name]
				ch <- Config{
					Release: release,
					ISO: []Source{
						webSource(f.URL, checksum, "", f.Name),
					},
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
