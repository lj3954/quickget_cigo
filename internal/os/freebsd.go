package os

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	freebsdX86Mirror     = "https://download.freebsd.org/ftp/releases/amd64/amd64/"
	freebsdAarch64Mirror = "https://download.freebsd.org/ftp/releases/arm64/aarch64/"
	freebsdRiscv64Mirror = "https://download.freebsd.org/ftp/releases/riscv/riscv64/"
)

type FreeBSD struct{}

func (FreeBSD) Data() OSData {
	return OSData{
		Name:        "freebsd",
		PrettyName:  "FreeBSD",
		Homepage:    "https://www.freebsd.org/",
		Description: "Operating system used to power modern servers, desktops, and embedded platforms.",
	}
}

func (FreeBSD) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	ch, wg := getChannelsWith(3)
	releaseRe := regexp.MustCompile(`href="([0-9\.]+)-RELEASE`)
	go buildFreeBSDConfigs(freebsdX86Mirror, "amd64", x86_64, ch, wg, errs, csErrs, releaseRe)
	go buildFreeBSDConfigs(freebsdX86Mirror, "arm64-aarch64", aarch64, ch, wg, errs, csErrs, releaseRe)
	go buildFreeBSDConfigs(freebsdRiscv64Mirror, "riscv-riscv64", riscv64, ch, wg, errs, csErrs, releaseRe)

	return waitForConfigs(ch, wg), nil
}

func buildFreeBSDConfigs(url, denom string, arch Arch, ch chan Config, wg *sync.WaitGroup, errs, csErrs chan<- Failure, releaseRe *regexp.Regexp) {
	defer wg.Done()
	releases, numReleases, err := getBasicReleases(url, releaseRe, -1)
	if err != nil {
		errs <- Failure{Error: err}
		return
	}
	wg.Add(2 * numReleases)

	freebsdEditions := [2]string{"disc1", "dvd1"}
	for release := range releases {
		go func() {
			defer wg.Done()
			checksumUrl := fmt.Sprintf("%sISO-IMAGES/%s/CHECKSUM.SHA256-FreeBSD-%s-RELEASE-%s", url, release, release, denom)
			checksums, err := cs.Build(cs.Sha256Regex, checksumUrl)
			if err != nil {
				csErrs <- Failure{Error: err}
			}
			for _, edition := range freebsdEditions {
				iso := fmt.Sprintf("FreeBSD-%s-RELEASE-%s-%s.iso.xz", release, denom, edition)
				checksum := checksums[iso]
				url := fmt.Sprintf("%sISO-IMAGES/%s/%s", url, release, iso)
				ch <- Config{
					Release: release,
					Edition: edition,
					GuestOS: quickgetdata.FreeBSD,
					Arch:    arch,
					ISO: []Source{
						webSource(url, checksum, quickgetdata.Xz, ""),
					},
				}
			}
		}()

		// VM Images (qcow2)
		go func() {
			defer wg.Done()
			mirrorArch := arch
			if arch == x86_64 {
				mirrorArch = "amd64"
			}
			mirror := fmt.Sprintf("https://download.freebsd.org/ftp/releases/VM-IMAGES/%s-RELEASE/%s/Latest/", release, mirrorArch)
			iso := fmt.Sprintf("FreeBSD-%s-RELEASE-%s.qcow2.xz", release, denom)
			checksumUrl := mirror + "CHECKSUM.SHA256"
			checksums, err := cs.Build(cs.Sha256Regex, checksumUrl)
			if err != nil {
				csErrs <- Failure{Error: err}
			}
			checksum := checksums[iso]
			url := mirror + iso
			ch <- Config{
				Release: release,
				Edition: "vm-image",
				GuestOS: quickgetdata.FreeBSD,
				Arch:    arch,
				DiskImages: []Disk{
					{
						Source: webSource(url, checksum, quickgetdata.Xz, ""),
					},
				},
			}
		}()
	}
}
