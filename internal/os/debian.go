package os

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
)

const (
	latestDebianMirror = "https://cdimage.debian.org/debian-cd/"
	prevDebianMirror   = "https://cdimage.debian.org/cdimage/archive/"
)

var (
	debianReleaseRe = regexp.MustCompile(`href="([0-9.]+)/"`)
	debianLiveRe    = regexp.MustCompile(`>(debian-live-[0-9.]+-amd64-([^.]+).iso)<`)
	debianNetinstRe = regexp.MustCompile(`>(debian-[0-9].+-(?:amd64|arm64)-(netinst).iso)<`)
)

type Debian struct{}

func (Debian) Data() OSData {
	return OSData{
		Name:        "debian",
		PrettyName:  "Debian",
		Homepage:    "https://www.debian.org/",
		Description: "Complete Free Operating System with perfect level of ease of use and stability.",
	}
}

func (Debian) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	ch, wg := getChannels()

	latestRelease := getLatestDebianConfigs(ch, &wg, errs, csErrs)
	getOldDebianConfigs(ch, &wg, errs, csErrs, latestRelease)

	return waitForConfigs(ch, &wg), nil
}

func getLatestDebianConfigs(ch chan Config, wg *sync.WaitGroup, errs, csErrs chan Failure) int {
	page, err := capturePage(latestDebianMirror)
	if err != nil {
		errs <- Failure{Error: err}
		return 0
	}

	fullRelease := debianReleaseRe.FindStringSubmatch(page)[1]
	dotIndex := strings.Index(fullRelease, ".")
	release := fullRelease[:dotIndex]
	latestRelease, err := strconv.Atoi(release)
	if err != nil {
		errs <- Failure{Error: err}
	}

	addConfigs(latestDebianMirror, release, fullRelease, ch, wg, errs, csErrs)
	return latestRelease
}

func getOldDebianConfigs(ch chan Config, wg *sync.WaitGroup, errs, csErrs chan Failure, latestRelease int) {
	page, err := capturePage(prevDebianMirror)
	if err != nil {
		errs <- Failure{Error: err}
		return
	}
	releaseMap := createReleaseMap(page, errs)
	if latestRelease == 0 {
		latestRelease = slices.Max(slices.Collect(maps.Keys(releaseMap))) + 1
	}

	for release := latestRelease - 2; release < latestRelease; release++ {
		addConfigs(prevDebianMirror, strconv.Itoa(release), releaseMap[release], ch, wg, errs, csErrs)
	}
}

func createReleaseMap(html string, errs chan Failure) map[int]string {
	m := make(map[int]string)
	for _, match := range debianReleaseRe.FindAllStringSubmatch(html, -1) {
		fullRelease := match[1]
		fullSemver, err := version.NewVersion(fullRelease)
		if err != nil {
			continue
		}
		dotIndex := strings.Index(fullRelease, ".")
		release, err := strconv.Atoi(match[1][:dotIndex])
		if err != nil {
			errs <- Failure{Error: err}
			continue
		}
		if prev, err := version.NewVersion(m[release]); err != nil || fullSemver.Compare(prev) > 0 {
			m[release] = fullRelease
		}
	}
	return m
}

func addConfigs(mirror, release, fullRelease string, ch chan Config, wg *sync.WaitGroup, errs, csErrs chan Failure) {
	liveMirror := mirror + fullRelease + "-live/amd64/iso-hybrid/"

	wg.Add(1)
	go func() {
		defer wg.Done()
		page, err := capturePage(liveMirror)
		if err != nil {
			errs <- Failure{Release: release, Error: err}
			return
		}
		checksums, err := buildChecksum(Whitespace{}, liveMirror+"SHA256SUMS")
		if err != nil {
			csErrs <- Failure{Release: release, Error: err}
		}
		for _, match := range debianLiveRe.FindAllStringSubmatch(page, -1) {
			iso := match[1]
			url := liveMirror + iso
			checksum := checksums[iso]
			ch <- Config{
				Release: release,
				Edition: match[2],
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}
	}()

	architectures := [2]string{"amd64", "arm64"}
	wg.Add(len(architectures))
	for _, arch := range architectures {
		netInstMirror := fmt.Sprintf("%s%s/%s/iso-cd/", mirror, fullRelease, arch)
		go func() {
			defer wg.Done()
			page, err := capturePage(netInstMirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			checksums, err := buildChecksum(Whitespace{}, netInstMirror+"SHA256SUMS")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			for _, match := range debianNetinstRe.FindAllStringSubmatch(page, -1) {
				iso := match[1]
				url := netInstMirror + iso
				checksum := checksums[iso]
				ch <- Config{
					Release: release,
					Edition: match[2],
					Arch:    Arch(arch),
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}
		}()
	}
}