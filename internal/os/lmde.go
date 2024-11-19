package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const lmdeMirror = "https://mirrors.edge.kernel.org/linuxmint/debian/"

type LMDE struct{}

func (LMDE) Data() OSData {
	return OSData{
		Name:        "lmde",
		PrettyName:  "Linux Mint Debian Edition",
		Homepage:    "https://linuxmint.com/download_lmde.php",
		Description: "Aims to be as similar as possible to Linux Mint, but without using Ubuntu. The package base is provided by Debian instead.",
	}
}

func (LMDE) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := capturePage(lmdeMirror)
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
