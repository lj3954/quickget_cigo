package os

import (
	"fmt"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	rockyMirror    = "https://dl.rockylinux.org/vault/rocky/"
	rockyReleaseRe = `href="(\d+\.\d+)/"`
)

var rockyLinux = OS{
	Name:           "rockylinux",
	PrettyName:     "Rocky Linux",
	Homepage:       "https://rockylinux.org/",
	Description:    "Open-source enterprise operating system designed to be 100% bug-for-bug compatible with Red Hat Enterprise Linux®.",
	ConfigFunction: createRockyLinuxConfigs,
}

func createRockyLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(rockyMirror, rockyReleaseRe, 3)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases * len(x86_64_aarch64))
	editions := [...]string{"boot", "dvd", "minimal"}

	for release := range releases {
		for _, arch := range x86_64_aarch64 {
			go func() {
				defer wg.Done()
				url := rockyMirror + release + "/isos/" + string(arch) + "/"

				checksums, err := cs.Build(cs.Sha256Regex, url+"CHECKSUM")
				if err != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: err}
				}

				for _, edition := range editions {
					iso := fmt.Sprintf("Rocky-%s-%s-%s.iso", release, arch, edition)
					url := url + iso
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
	return waitForConfigs(ch, wg), nil
}
