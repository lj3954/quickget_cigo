package os

import (
	"fmt"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	opensuseLeapMirror = "https://download.opensuse.org/distribution/leap/"
	opensuseReleaseRe  = `href=".\/(\d{2}\.\d{1})`
)

var OpenSUSE = OS{
	Name:           "opensuse",
	PrettyName:     "openSUSE",
	Homepage:       "https://www.opensuse.org/",
	Description:    "The makers choice for sysadmins, developers and desktop users.",
	ConfigFunction: createOpenSUSEConfigs,
}

func createOpenSUSEConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, _, err := getReverseReleases(opensuseLeapMirror, opensuseReleaseRe, 5)
	if err != nil {
		return nil, err
	}
	architectures := []Arch{x86_64, aarch64}
	ch, wg := getChannels()
	for release := range releases {
		if release == "42.3" {
			continue
		}
		wg.Add(len(architectures))
		for _, arch := range architectures {
			go func() {
				defer wg.Done()
				iso := fmt.Sprintf("openSUSE-Leap-%s-DVD-x86_64-Current.iso", release)
				url := fmt.Sprintf("%s%s/iso/%s", opensuseLeapMirror, release, iso)
				checksum, err := cs.SingleWhitespace(url + ".sha256")
				if err != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: err}
				}
				ch <- Config{
					Release: release,
					Arch:    arch,
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}()
		}
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		tumbleweedUrl := "https://download.opensuse.org/tumbleweed/iso/openSUSE-Tumbleweed-DVD-x86_64-Current.iso"
		checksum, err := cs.SingleWhitespace(tumbleweedUrl + ".sha256")
		if err != nil {
			csErrs <- Failure{Release: "tumbleweed", Arch: x86_64, Error: err}
		}
		ch <- Config{
			Release: "tumbleweed",
			Arch:    x86_64,
			ISO: []Source{
				urlChecksumSource(tumbleweedUrl, checksum),
			},
		}
	}()

	go func() {
		defer wg.Done()
		microOSUrl := "https://download.opensuse.org/tumbleweed/iso/openSUSE-MicroOS-DVD-x86_64-Current.iso"
		checksum, err := cs.SingleWhitespace(microOSUrl + ".sha256")
		if err != nil {
			csErrs <- Failure{Release: "microos", Arch: x86_64, Error: err}
		}
		ch <- Config{
			Release: "microos",
			Arch:    x86_64,
			ISO: []Source{
				urlChecksumSource(microOSUrl, checksum),
			},
		}
	}()

	go func() {
		defer wg.Done()
		aeonUrl := "https://mirrorcache.opensuse.org/tumbleweed/appliances/iso/opensuse-aeon.x86_64.iso"
		checksum, err := cs.SingleWhitespace(aeonUrl + ".sha256")
		if err != nil {
			csErrs <- Failure{Release: "aeon", Arch: x86_64, Error: err}
		}
		ch <- Config{
			Release: "aeon",
			Arch:    x86_64,
			ISO: []Source{
				urlChecksumSource(aeonUrl, checksum),
			},
		}
	}()

	return waitForConfigs(ch, wg), nil
}
