package os

import (
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const (
	openbsdMirror    = "https://mirror.leaseweb.com/pub/OpenBSD/"
	openbsdReleaseRe = `href="(\d[\d\.]+)/"`
)

type OpenBSD struct{}

func (OpenBSD) Data() OSData {
	return OSData{
		Name:        "openbsd",
		PrettyName:  "OpenBSD",
		Homepage:    "https://www.openbsd.org/",
		Description: "FREE, multi-platform 4.4BSD-based UNIX-like operating system. Our efforts emphasize portability, standardization, correctness, proactive security and integrated cryptography.",
	}
}

func (OpenBSD) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(openbsdMirror, openbsdReleaseRe, 4)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(3 * numReleases)

	for release := range releases {
		amd64Mirror := openbsdMirror + release + "/amd64/"
		go addOpenBSDConfig(amd64Mirror, release, x86_64, ch, wg, csErrs)
		arm64Mirror := openbsdMirror + release + "/arm64/"
		go addOpenBSDConfig(arm64Mirror, release, aarch64, ch, wg, csErrs)
		riscv64Mirror := openbsdMirror + release + "/riscv64/"
		go addOpenBSDConfig(riscv64Mirror, release, riscv64, ch, wg, csErrs)
	}

	return waitForConfigs(ch, wg), nil
}

func addOpenBSDConfig(mirror, release string, arch Arch, ch chan Config, wg *sync.WaitGroup, csErrs chan Failure) {
	defer wg.Done()

	checksums, err := cs.Build(cs.Sha256Regex, mirror+"SHA256")
	if err != nil {
		csErrs <- Failure{Release: release, Arch: arch, Error: err}
	}
	iBase := "install" + strings.ReplaceAll(release, ".", "")

	config := Config{
		GuestOS: quickgetdata.GenericBSD,
		Release: release,
		Arch:    arch,
	}

	switch arch {
	case riscv64:
		img := iBase + ".img"
		url := mirror + img
		checksum := checksums[img]
		config.IMG = []Source{
			urlChecksumSource(url, checksum),
		}
	default:
		iso := iBase + ".iso"
		url := mirror + iso
		checksum := checksums[iso]
		config.ISO = []Source{
			urlChecksumSource(url, checksum),
		}
	}
	ch <- config
}
