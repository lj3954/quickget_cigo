package os

import (
	"fmt"
	"log"
	"regexp"
	"strings"
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
	errs := make(chan error)
	var wg sync.WaitGroup
	isoRegex := regexp.MustCompile(`<a href="(AlmaLinux-[0-9]+-latest-(?:x86_64|aarch64)-([^-]+).iso)">`)

	architectures := [2]Arch{x86_64, aarch64}
	for _, release := range releases {
		for _, arch := range architectures {
			mirror := fmt.Sprintf("%s%s/isos/%s", AlmaMirror, release, arch)
			wg.Add(1)
			go func() {
				defer wg.Done()
				page, err := capturePage(mirror)
				if err != nil {
					errs <- err
					return
				}
				checksums, err := buildChecksum(Whitespace{}, mirror+"CHECKSUM")
				if err != nil {
					errs <- err
				}
				for _, match := range isoRegex.FindAllStringSubmatch(page, -1) {
					if strings.HasSuffix(match[0], ".manifest") {
						continue
					}
					iso, edition := match[1], match[2]
					url := mirror + iso
					checksum := checksums[iso]
					ch <- Config{
						Release: release,
						Edition: edition,
						Arch:    arch,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				}
			}()
		}
	}

	go func() {
		wg.Wait()
		close(ch)
		close(errs)
	}()
	for err := range errs {
		log.Println(err)
	}
	configs := make([]Config, 0)
	for config := range ch {
		configs = append(configs, config)
	}
	return configs, nil
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
