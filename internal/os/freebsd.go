package os

import (
	"fmt"
	"sync"

	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const (
	freebsdX86Mirror     = "https://download.freebsd.org/ftp/releases/amd64/amd64/"
	freebsdAarch64Mirror = "https://download.freebsd.org/ftp/releases/arm64/aarch64/"
	freebsdRiscv64Mirror = "https://download.freebsd.org/ftp/releases/riscv/riscv64/"
	freebsdReleaseRe     = `href="([0-9\.]+)-RELEASE`
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

func (FreeBSD) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	ch, wg := getChannels()
	wg.Add(3)
	go buildFreeBSDConfigs(freebsdX86Mirror, "amd64", x86_64, ch, &wg, errs, csErrs)
	go buildFreeBSDConfigs(freebsdX86Mirror, "arm64-aarch64", aarch64, ch, &wg, errs, csErrs)
	go buildFreeBSDConfigs(freebsdRiscv64Mirror, "riscv-riscv64", riscv64, ch, &wg, errs, csErrs)

	return waitForConfigs(ch, &wg), nil
}

func buildFreeBSDConfigs(url, denom string, arch Arch, ch chan Config, wg *sync.WaitGroup, errs, csErrs chan Failure) {
	defer wg.Done()
	releases, err := getBasicReleases(url, freebsdReleaseRe, -1)
	if err != nil {
		errs <- Failure{Error: err}
		return
	}

	freebsdEditions := [2]string{"disc1", "dvd1"}
	for _, release := range releases {
		wg.Add(2)
		go func() {
			defer wg.Done()
			checksumUrl := fmt.Sprintf("%sISO-IMAGES/%s/CHECKSUM.SHA256-FreeBSD-%s-RELEASE-%s", url, release, release, denom)
			checksums, err := buildChecksum(Sha256Regex, checksumUrl)
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
			checksums, err := buildChecksum(Sha256Regex, checksumUrl)
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
