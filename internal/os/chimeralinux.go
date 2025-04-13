package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const chimeraMirror = "https://repo.chimera-linux.org/live/latest/"

var chimeraLinux = OS{
	Name:           "chimeralinux",
	PrettyName:     "Chimera Linux",
	Homepage:       "https://chimera-linux.org/",
	Description:    "Modern, general-purpose non-GNU Linux distribution.",
	ConfigFunction: createChimeraLinuxConfigs,
}

func createChimeraLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(chimeraMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="(chimera-linux-(x86_64|aarch64|riscv64)-LIVE-[0-9]{8}-([^-]+).iso)"`)

	checksums, err := cs.Build(cs.Whitespace{}, chimeraMirror+"sha256sums.txt")
	if err != nil {
		csErrs <- Failure{Release: "latest", Error: err}
	}

	configs := make([]Config, 0)
	for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
		checksum := checksums[match[1]]
		url := chimeraMirror + match[1]
		configs = append(configs, Config{
			Release: "latest",
			Edition: match[3],
			Arch:    Arch(match[2]),
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		})
	}
	return configs, nil
}
