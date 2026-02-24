package os

import (
	"errors"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const nitruxMirror = "https://sourceforge.net/projects/nitruxos/files/Release/"

var Nitrux = OS{
	Name:           "nitrux",
	PrettyName:     "Nitrux",
	Homepage:       "https://nxos.org/",
	Description:    "Powered by Debian, KDE Plasma and Frameworks, and AppImages.",
	ConfigFunction: createNitruxConfigs,
}

func createNitruxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(nitruxMirror)
	if err != nil {
		return nil, err
	}

	release := "latest"

	isoSubDir, ok := head.SubDirs["ISO"]
	if !ok {
		return nil, errors.New("iso directory doesn't exist")
	}
	isoDir, err := isoSubDir.Fetch()
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`^nitrux-contemporary-(.*?)-[0-9a-f]{8}-([^\.]+)\.iso$`)

	var checksumDir *mirror.Directory
	if d, ok := head.SubDirs["SHA512"]; ok {
		checksumDir, err = d.Fetch()
		if err != nil {
			csErrs <- Failure{Release: release, Error: err}
		}
	}

	var configs []Config
	for f, match := range isoDir.FileMatches(isoRe) {
		edition := match[1]
		arch, v := NewArch(match[2])
		if !v {
			continue
		}

		var checksum string
		checksumName := strings.TrimSuffix(f.Name, ".iso") + ".sha512"
		if checksumDir != nil {
			if cf, ok := checksumDir.Files[checksumName]; ok {
				checksum, err = cs.SingleWhitespace(cf)
				if err != nil {
					csErrs <- Failure{Release: release, Edition: edition, Arch: arch, Error: err}
				}
			}
		}

		configs = append(configs, Config{
			Release: release,
			Edition: edition,
			Arch:    arch,
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		})
	}

	return configs, nil
}
