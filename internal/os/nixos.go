package os

import (
	"fmt"
	"iter"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	nixDataUrl     = "https://nix-channels.s3.amazonaws.com/?delimiter=/"
	nixDownloadUrl = "https://channels.nixos.org"
)

type NixOS struct{}

func (NixOS) Data() OSData {
	return OSData{
		Name:        "nixos",
		PrettyName:  "NixOS",
		Homepage:    "https://nixos.org/",
		Description: "Linux distribution based on Nix package manager, tool that takes a unique approach to package management and system configuration.",
	}
}

func (NixOS) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getNixReleases(6)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`latest-nixos-([^-]+)-(x86_64|aarch64)-linux.iso`)
	ch, wg := getChannels()

	for release := range releases {
		mirror := fmt.Sprintf("%s&prefix=nixos-%s/", nixDataUrl, release)
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := getNixXML(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			for _, entry := range data.Contents {
				if !strings.HasSuffix(entry.Key, ".iso") {
					continue
				}
				match := isoRe.FindStringSubmatch(entry.Key)
				if match == nil {
					continue
				}
				name, edition, arch := match[0], match[1], Arch(match[2])
				url := fmt.Sprintf("%s/nixos-%s/%s", nixDownloadUrl, release, name)

				wg.Add(1)
				go func() {
					defer wg.Done()
					checksum, err := cs.SingleWhitespace(url + ".sha256")
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Arch: arch, Error: err}
					}
					ch <- Config{
						Release: release,
						Edition: edition,
						Arch:    arch,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				}()
			}
		}()
	}
	return waitForConfigs(ch, &wg), nil
}

func getNixXML(url string) (*nixReleases, error) {
	var releaseData nixReleases
	if err := capturePageToXml(url, &releaseData); err != nil {
		return nil, err
	}
	return &releaseData, nil
}

func getNixReleases(count int) (iter.Seq[string], error) {
	releaseData, err := getNixXML(nixDataUrl)
	if err != nil {
		return nil, err
	}
	return func(yield func(string) bool) {
		releaseRe := regexp.MustCompile(`nixos-(([0-9]+.[0-9]+|(unstable))(?:-small)?)`)
		for i := len(releaseData.Contents) - 1; i >= 0 && count > 0; i-- {
			if match := releaseRe.FindStringSubmatch(releaseData.Contents[i].Key); match != nil {
				if !yield(match[1]) {
					return
				}
				count--
			}
		}
	}, nil
}

type nixReleases struct {
	Contents []struct {
		Key string
	}
}
