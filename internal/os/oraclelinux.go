package os

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	oracleLinuxChecksumMirror = "https://linux.oracle.com/security/gpg/checksum/"
	oracleLinuxReleaseRe      = `href="(OracleLinux-R(\d+)-U(\d+)-Server-([^\.]+)\.checksum)`
)

type OracleLinux struct{}

func (OracleLinux) Data() OSData {
	return OSData{
		Name:        "oraclelinux",
		PrettyName:  "Oracle Linux",
		Homepage:    "https://www.oracle.com/linux/",
		Description: "Linux with everything required to deploy, optimize, and manage applications on-premises, in the cloud, and at the edge.",
	}
}

func (OracleLinux) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(oracleLinuxChecksumMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(oracleLinuxReleaseRe)
	releases := releaseRe.FindAllStringSubmatch(page, -1)
	slices.SortFunc(releases, func(a, b []string) int {
		if a[2] < b[2] {
			return 1
		} else if a[2] > b[2] {
			return -1
		}
		if a[3] < b[3] {
			return 1
		} else if a[3] > b[3] {
			return -1
		}
		return 0
	})

	ch, wg := getChannels()
	for i, match := range releases {
		if i >= 4 {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			major, minor, arch := match[2], match[3], match[4]
			release := major + "." + minor
			checksumData, err := web.CapturePage(oracleLinuxChecksumMirror + match[1])
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			for _, line := range strings.Split(checksumData, "\n") {
				if !(strings.Contains(line, "dvd") && strings.Contains(line, "OracleLinux")) {
					continue
				}
				splitLine := strings.Fields(line)
				checksum := splitLine[0]
				iso := splitLine[1]
				url := fmt.Sprintf("https://yum.oracle.com/ISOS/OracleLinux/OL%s/u%s/%s/%s", major, minor, arch, iso)
				ch <- Config{
					Release: release,
					Arch:    Arch(arch),
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
