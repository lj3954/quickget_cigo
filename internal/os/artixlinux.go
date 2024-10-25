package os

import (
	"log"
	"regexp"
)

const ArtixMirror = "https://mirrors.ocf.berkeley.edu/artix-iso/"

type ArtixLinux struct{}

func (ArtixLinux) Data() OSData {
	return OSData{
		Name:        "artixlinux",
		PrettyName:  "Artix Linux",
		Homepage:    "https://artixlinux.org/",
		Description: "The Art of Linux. Simple. Fast. Systemd-free.",
	}
}

func (ArtixLinux) CreateConfigs() ([]Config, error) {
	page, err := capturePage(ArtixMirror)
	if err != nil {
		return nil, err
	}
	checksums, err := buildChecksum(Whitespace{}, ArtixMirror+"sha256sums")
	if err != nil {
		log.Println(err)
	}
	isoRe := regexp.MustCompile(`href="(artix-(.*?)-([^-]+-[0-9]+)-x86_64.iso)"`)

	matches := isoRe.FindAllStringSubmatch(page, -1)
	configs := make([]Config, len(matches))

	for i, match := range matches {
		iso, edition, release := match[1], match[2], match[3]
		url := ArtixMirror + iso
		checksum := checksums[iso]
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
