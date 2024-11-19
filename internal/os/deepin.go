package os

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const (
	deepinMirror    = "https://cdimage.deepin.com/releases/"
	deepinReleaseRe = `class="name">([\d.]+)\/`
)

type Deepin struct{}

func (Deepin) Data() OSData {
	return OSData{
		Name:        "deepin",
		PrettyName:  "Deepin",
		Homepage:    "https://www.deepin.org/",
		Description: "Beautiful UI design, intimate human-computer interaction, and friendly community environment make you feel at home.",
	}
}

func (Deepin) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, numReleases, err := getBasicReleases(deepinMirror, deepinReleaseRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numReleases)
	archRe := regexp.MustCompile(`class="name">(amd64|arm64)\/`)
	for release := range releases {
		mirror := deepinMirror + release + "/"
		go func() {
			defer wg.Done()
			architectures, numArchitectures, err := getBasicReleases(mirror, archRe, -1)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			if numArchitectures > 0 {
				wg.Add(numArchitectures)
				for arch := range architectures {
					mirror := mirror + arch + "/"
					go addDeepinConfigs(mirror, release, Arch(arch), ch, wg, csErrs)
				}
			} else {
				wg.Add(1)
				go addDeepinConfigs(mirror, release, x86_64, ch, wg, csErrs)
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}

func addDeepinConfigs(url, release string, arch Arch, ch chan Config, wg *sync.WaitGroup, csErrs chan Failure) {
	defer wg.Done()
	isoUrl := fmt.Sprintf("%sdeepin-desktop-community-%s-%s.iso", url, release, arch)
	checksum, err := cs.SingleWhitespace(url + "SHA256SUMS")
	if err != nil {
		csErrs <- Failure{Release: release, Error: err}
		return
	}
	ch <- Config{
		Release: release,
		Arch:    arch,
		ISO: []Source{
			urlChecksumSource(isoUrl, checksum),
		},
	}
}
