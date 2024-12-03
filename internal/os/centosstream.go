package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	centOSMirror    = "https://linuxsoft.cern.ch/centos-stream/"
	centOSUrlFormat = "https://mirrors.centos.org/mirrorlist?path=/%s%s&redirect=1&protocol=https"
	centOSReleaseRe = `href="([0-9]+)-stream/"`
)

type CentOSStream struct{}

func (CentOSStream) Data() OSData {
	return OSData{
		Name:        "centos-stream",
		PrettyName:  "CentOS Stream",
		Homepage:    "https://www.centos.org/centos-stream/",
		Description: "Continuously delivered distro that tracks just ahead of Red Hat Enterprise Linux (RHEL) development, positioned as a midstream between Fedora Linux and RHEL.",
	}
}

func (CentOSStream) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(centOSMirror, centOSReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="(CentOS-Stream-[0-9]+-[0-9]{8}.0-[^-]+-([^-]+)\.iso)"`)
	ch, wg := getChannelsWith(numReleases * len(x86_64_aarch64))
	for release := range releases {
		for _, arch := range x86_64_aarch64 {
			mirrorAdd := fmt.Sprintf("%s-stream/BaseOS/%s/iso/", release, arch)
			mirror := centOSMirror + mirrorAdd

			go func() {
				defer wg.Done()
				page, err := web.CapturePage(mirror)
				if err != nil {
					errs <- Failure{Release: release, Arch: arch, Error: err}
					return
				}
				checksums, err := cs.Build(cs.Sha256Regex, mirror+"SHA256SUM")
				if err != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: err}
				}
				for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
					iso := match[1]
					url := fmt.Sprintf(centOSUrlFormat, mirrorAdd, iso)
					checksum := checksums[iso]
					ch <- Config{
						Release: release,
						Edition: match[2],
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
