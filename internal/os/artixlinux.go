package os

import (
	"regexp"
)

const artixMirror = "https://mirrors.ocf.berkeley.edu/artix-iso/"

type ArtixLinux struct{}

func (ArtixLinux) Data() OSData {
	return OSData{
		Name:        "artixlinux",
		PrettyName:  "Artix Linux",
		Homepage:    "https://artixlinux.org/",
		Description: "The Art of Linux. Simple. Fast. Systemd-free.",
	}
}

func (ArtixLinux) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := capturePage(artixMirror)
	if err != nil {
		return nil, err
	}
	checksums, err := buildChecksum(Whitespace{}, artixMirror+"sha256sums")
	if err != nil {
		csErrs <- Failure{Error: err}
	}
	isoRe := regexp.MustCompile(`href="(artix-(.*?)-([^-]+-[0-9]+)-x86_64.iso)"`)

	matches := isoRe.FindAllStringSubmatch(page, -1)
	configs := make([]Config, len(matches))

	for i, match := range matches {
		iso, edition, release := match[1], match[2], match[3]
		url := artixMirror + iso
		checksum, _ := checksums[iso]
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
