package os

import (
	"fmt"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	netbsdMirror           = "https://cdn.netbsd.org/pub/NetBSD/iso/"
	netbsdReleaseRe        = `href="([0-9]+\.[0-9]+)/"`
	netbsdAmd64IsoFormat   = "NetBSD-%s-amd64.iso"
	netbsdAarch64IsoFormat = "NetBSD-%s-evbarm-aarch64.iso"
)

type NetBSD struct{}

func (NetBSD) Data() OSData {
	return OSData{
		Name:        "netbsd",
		PrettyName:  "NetBSD",
		Homepage:    "https://www.netbsd.org/",
		Description: "Free, fast, secure, and highly portable Unix-like Open Source operating system. It is available for a wide range of platforms, from large-scale servers and powerful desktop systems to handheld and embedded devices.",
	}
}

func (NetBSD) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, err := getSortedReleasesFunc(netbsdMirror, netbsdReleaseRe, 4, semverCompare)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(len(releases))
	for _, release := range releases {
		mirror := netbsdMirror + release + "/"
		go func() {
			defer wg.Done()
			checksums, err := cs.Build(cs.Sha512Regex, mirror+"SHA512")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
			ch <- getNetBSDConfig(checksums, mirror, fmt.Sprintf(netbsdAmd64IsoFormat, release), release, x86_64)
			ch <- getNetBSDConfig(checksums, mirror, fmt.Sprintf(netbsdAarch64IsoFormat, release), release, aarch64)
		}()
	}
	return waitForConfigs(ch, wg), nil
}

func getNetBSDConfig(checksums map[string]string, mirror, iso, release string, arch Arch) Config {
	url := mirror + iso
	checksum := checksums[iso]

	return Config{
		GuestOS: quickgetdata.GenericBSD,
		Release: release,
		Arch:    arch,
		ISO: []Source{
			urlChecksumSource(url, checksum),
		},
	}
}
