package os

import (
	"regexp"
	"slices"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const kaliMirror = "https://cdimage.kali.org/"

type Kali struct{}

func (Kali) Data() OSData {
	return OSData{
		Name:        "kali",
		PrettyName:  "Kali Linux",
		Homepage:    "https://www.kali.org/",
		Description: "The most advanced Penetration Testing Distribution.",
	}
}

func (Kali) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases := [...]string{"current", "kali-weekly"}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`href="(kali-linux-\d{4}-[^-]+-(installer|live)-(amd64|arm64).iso)"`)
	wg.Add(len(releases))
	for _, release := range releases {
		mirror := kaliMirror + release + "/"
		go func() {
			defer wg.Done()
			matches, err := getKaliMatches(mirror, isoRe)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			checksums, err := cs.Build(cs.Whitespace{}, mirror+"SHA256SUMS")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			for _, match := range matches {
				iso, edition, arch := match[1], match[2], Arch(match[3])
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

	return waitForConfigs(ch, &wg), nil
}

func getKaliMatches(url string, isoRe *regexp.Regexp) ([][]string, error) {
	page, err := capturePage(url)
	if err != nil {
		return nil, err
	}
	matches := isoRe.FindAllStringSubmatch(page, -1)

	slices.Reverse(matches)
	set := make(map[string]struct{})
	return slices.DeleteFunc(matches, func(match []string) bool {
		if _, ok := set[match[2]+match[3]]; ok {
			return true
		}
		set[match[2]+match[3]] = struct{}{}
		return false
	}), nil
}
