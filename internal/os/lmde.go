package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const lmdeMirror = "https://mirrors.edge.kernel.org/linuxmint/debian/"

var LMDE = OS{
	Name:           "lmde",
	PrettyName:     "Linux Mint Debian Edition",
	Homepage:       "https://linuxmint.com/download_lmde.php",
	Description:    "Aims to be as similar as possible to Linux Mint, but without using Ubuntu. The package base is provided by Debian instead.",
	ConfigFunction: createLmdeConfigs,
}

func createLmdeConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(lmdeMirror)
	if err != nil {
		return nil, err
	}
	checksums, err := cs.Build(cs.Whitespace{}, lmdeMirror+"sha256sum.txt")
	if err != nil {
		csErrs <- Failure{Error: err}
	}
	isoRe := regexp.MustCompile(`href="(lmde-(\d+(?:\.\d+)?)-(\w+)-64bit.iso)"`)

	matches := isoRe.FindAllStringSubmatch(page, -1)
	configs := make([]Config, len(matches))

	for i, match := range isoRe.FindAllStringSubmatch(page, -1) {
		iso, release, edition := match[1], match[2], match[3]
		checksum := checksums["*"+iso]
		url := lmdeMirror + iso
		configs[i] = Config{
			Release: release,
			Edition: edition,
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		}
	}
	return configs, nil
}
