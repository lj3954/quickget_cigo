package os

import (
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const dragonflybsdMirror = "https://mirror-master.dragonflybsd.org/iso-images/"

type DragonFlyBSD struct{}

func (DragonFlyBSD) Data() OSData {
	return OSData{
		Name:        "dragonflybsd",
		PrettyName:  "DragonFlyBSD",
		Homepage:    "https://www.dragonflybsd.org/",
		Description: "Provides an opportunity for the BSD base to grow in an entirely different direction from the one taken in the FreeBSD, NetBSD, and OpenBSD series.",
	}
}

func (DragonFlyBSD) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(dragonflybsdMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="(dfly-x86_64-([0-9.]+)_REL.iso.bz2)"`)
	checksums, err := cs.Build(cs.Md5Regex, dragonflybsdMirror+"md5.txt")
	if err != nil {
		csErrs <- Failure{Error: err}
	}

	matches := isoRe.FindAllStringSubmatch(page, -1)

	// Remove duplicate values, ignoring patch releases
	matches = slices.CompactFunc(matches, func(a, b []string) bool {
		a = strings.SplitN(a[2], ".", 3)
		b = strings.SplitN(b[2], ".", 3)
		return a[0] == b[0] && a[1] == b[1]
	})

	numConfigs := min(len(matches), 4)
	configs := make([]Config, numConfigs)
	for i := range numConfigs {
		iso, release := matches[i][1], matches[i][2]
		url := dragonflybsdMirror + iso
		checksum := checksums[iso]
		configs[i] = Config{
			GuestOS: quickgetdata.GenericBSD,
			Release: release,
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		}
	}

	return configs, nil
}
