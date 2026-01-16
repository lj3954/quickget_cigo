package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const chimeraMirror = "https://repo.chimera-linux.org/live/latest/"

var ChimeraLinux = OS{
	Name:           "chimeralinux",
	PrettyName:     "Chimera Linux",
	Homepage:       "https://chimera-linux.org/",
	Description:    "Modern, general-purpose non-GNU Linux distribution.",
	ConfigFunction: createChimeraLinuxConfigs,
}

func createChimeraLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(chimeraMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`^chimera-linux-(x86_64|aarch64|riscv64)-LIVE-[0-9]{8}-([^-]+).iso$`)

	checksums := make(map[string]string)
	if f, e := head.Files["sha256sums.txt"]; e {
		checksums, err = cs.Build(cs.Whitespace, f)
		if err != nil {
			csErrs <- Failure{Release: "latest", Error: err}
		}
	}

	configs := make([]Config, 0)
	for f, match := range head.FileMatches(isoRe) {
		checksum := checksums[f.Name]
		configs = append(configs, Config{
			Release: "latest",
			Edition: match[2],
			Arch:    Arch(match[1]),
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		})

	}
	return configs, nil
}
