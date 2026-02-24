package os

import (
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	parrotSecMirror = "https://download.parrot.sh/parrot/iso/"
	parrotSecIsoRe  = `^Parrot-([^-]+)-[\d\.]+_([^\.]+)\.(iso|qcow2.xz)$`
)

var ParrotSec = OS{
	Name:           "parrotsec",
	PrettyName:     "Parrot Security",
	Homepage:       "https://www.parrotsec.org/",
	Description:    `Provides a huge arsenal of tools, utilities and libraries that IT and security professionals can use to test and assess the security of their assets in a reliable, compliant and reproducible way.`,
	ConfigFunction: createParrotSecConfigs,
}

func createParrotSecConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(parrotSecMirror)
	if err != nil {
		return nil, err
	}

	subdirs := head.NameSortedSubDirs(utils.SemverCompare)
	threeMostRecent := subdirs[max(len(subdirs)-3, 0):]

	ch, wg := getChannels()
	isoRe := regexp.MustCompile(parrotSecIsoRe)

	for _, releaseDir := range threeMostRecent {
		wg.Go(func() {
			release := releaseDir.Name
			contents, err := releaseDir.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			checksums := make(map[string]string)
			cf, ok := contents.FindFile(func(f mirror.File) bool {
				k := strings.ToLower(f.Name)
				return strings.HasSuffix(k, ".txt") && strings.Contains(k, "hash")
			})
			if ok {
				page, err := web.CapturePage(cf.URL)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				} else {
					lines := strings.Split(page, "\n")

					sha256Start := slices.Index(lines, "sha256")
					if sha256Start >= 0 {
						lines = lines[sha256Start:]
					}
					sha256End := slices.IndexFunc(lines, func(s string) bool {
						return len(strings.TrimSpace(s)) == 0
					})
					if sha256End > 0 {
						lines = lines[:sha256End]
					}

					contents := strings.Join(lines, "\n")
					checksums = cs.Whitespace.BuildWithData(contents)
				}
			}

			for f, match := range contents.FileMatches(isoRe) {
				arch, v := NewArch(match[2])
				if !v {
					continue
				}
				config := Config{
					Release: release,
					Edition: match[1],
					Arch:    arch,
				}

				qcowXz := match[3] == "qcow2.xz"
				if qcowXz {
					config.Edition += "-preinstalled"
				}

				checksum := checksums[f.Name]

				if qcowXz {
					config.DiskImages = []Disk{
						{
							Source: webSource(f.URL.String(), checksum, quickgetdata.Xz, f.Name),
						},
					}
				} else {
					config.ISO = []Source{
						webSource(f.URL.String(), checksum, "", f.Name),
					}
				}

				ch <- config
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
