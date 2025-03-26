package os

import (
	"fmt"
	"iter"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	almaMirror    = "https://repo.almalinux.org/almalinux/"
	almaReleaseRe = `<a href="([0-9]+)/"`
)

type Alma struct{}

func (Alma) Data() OSData {
	return OSData{
		Name:        "alma",
		PrettyName:  "AlmaLinux",
		Homepage:    "https://almalinux.org/",
		Description: "Community owned and governed, forever-free enterprise Linux distribution, focused on long-term stability, providing a robust production-grade platform. AlmaLinux OS is binary compatible with RHEL®.",
	}
}

func (Alma) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(almaMirror, almaReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases * len(x86_64_aarch64))
	isoRe := regexp.MustCompile(`<a href="(AlmaLinux-[0-9]+-latest-(?:x86_64|aarch64)-([^-]+).iso)">`)

	for release := range releases {
		for _, arch := range x86_64_aarch64 {
			go func() {
				defer wg.Done()
				configs, csErr, err := getAlmaConfigs(release, arch, isoRe)
				if err != nil {
					errs <- Failure{Release: release, Arch: arch, Error: err}
					return
				}
				if csErr != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: err}
				}
				for config := range configs {
					ch <- config
				}
			}()
		}
	}

	return waitForConfigs(ch, wg), nil
}

func getAlmaConfigs(release string, arch Arch, isoRe *regexp.Regexp) (configs iter.Seq[Config], csErr error, e error) {
	mirror := fmt.Sprintf("%s%s/isos/%s/", almaMirror, release, arch)
	page, err := web.CapturePage(mirror)
	if err != nil {
		return nil, nil, err
	}

	checksums, err := cs.Build(cs.Sha256Regex, mirror+"CHECKSUM")
	if err != nil {
		csErr = err
	}

	configs = func(yield func(Config) bool) {
		for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
			if strings.HasSuffix(match[0], ".manifest") {
				continue
			}

			iso, edition := match[1], match[2]
			url := mirror + iso
			checksum := checksums[iso]

			config := Config{
				Release: release,
				Edition: edition,
				Arch:    arch,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}

			if !yield(config) {
				break
			}
		}
	}

	return
}
