package os

import (
	"maps"
	"regexp"
	"slices"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const dragonflybsdMirror = "https://mirror-master.dragonflybsd.org/iso-images/"

var DragonFlyBSD = OS{
	Name:           "dragonflybsd",
	PrettyName:     "DragonFlyBSD",
	Homepage:       "https://www.dragonflybsd.org/",
	Description:    "Provides an opportunity for the BSD base to grow in an entirely different direction from the one taken in the FreeBSD, NetBSD, and OpenBSD series.",
	ConfigFunction: createDragonFlyBSDConfigs,
}

type dragonflybsdRelease struct {
	file    mirror.File
	patch   string
	release string
}

func createDragonFlyBSDConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(dragonflybsdMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`^dfly-x86_64-((\d+\.\d+)\.(\d+))_REL.iso.bz2$`)

	checksums := make(map[string]string)
	if f, e := head.Files["md5.txt"]; e {
		checksums, err = cs.Build(cs.Md5Regex, f)
		if err != nil {
			csErrs <- Failure{Error: err}
		}
	}

	releases := make(map[string]dragonflybsdRelease)
	for k, f := range head.Files {
		match := isoRe.FindStringSubmatch(k)
		if match == nil {
			continue
		}
		full, main, patch := match[1], match[2], match[3]
		if r, e := releases[main]; !e || patch > r.patch {
			releases[main] = dragonflybsdRelease{file: f, patch: patch, release: full}
		}
	}

	releaseSlice := slices.Collect(maps.Values(releases))
	slices.SortFunc(releaseSlice, func(a, b dragonflybsdRelease) int {
		return semverCompare(a.release, b.release)
	})

	latestFour := releaseSlice[max(len(releaseSlice)-4, 0):]
	configs := make([]Config, len(latestFour))
	for i, r := range latestFour {
		f := r.file
		checksum := checksums[f.Name]
		configs[i] = Config{
			GuestOS: quickgetdata.GenericBSD,
			Release: r.release,
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		}
	}

	return configs, nil
}
