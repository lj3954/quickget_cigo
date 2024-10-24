package os

import (
	"fmt"
	"regexp"
	"sync"
)

const AlmaMirror = "https://repo.almalinux.org/almalinux/"

type Alma struct{}

func (*Alma) Data() OSData {
	return OSData{
		Name:        "alma",
		PrettyName:  "AlmaLinux",
		Homepage:    "https://almalinux.org/",
		Description: "Community owned and governed, forever-free enterprise Linux distribution, focused on long-term stability, providing a robust production-grade platform. AlmaLinux OS is binary compatible with RHEL®.",
	}
}

func (*Alma) CreateConfigs() ([]Config, error) {
	releases, err := getReleases()
	if err != nil {
		return nil, err
	}
	ch := make(chan Config)
	var wg sync.WaitGroup
	isoRegex := regexp.MustCompile(`<a href="(AlmaLinux-[0-9]+-latest-(?:x86_64|aarch64)-([^-]+).iso)">`)

	architectures := [2]Arch{x86_64, aarch64}
	for _, release := range releases {
		for _, arch := range architectures {
			mirror := fmt.Sprintf("%s%s/isos/%s", AlmaMirror, release, arch)
			wg.Add(1)
			go func() {
				defer wg.Done()
			}
		}
	}
}

func getReleases() ([]string, error) {
	releaseHTML, err := capturePage(AlmaMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`<a href="([0-9]+)/"`)
	matches := releaseRe.FindAllStringSubmatch(releaseHTML, -1)

	releases := make([]string, len(matches))
	for i, match := range matches {
		releases[i] = match[1]
	}
	return releases, nil
}
