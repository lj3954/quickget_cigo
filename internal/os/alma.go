package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

const (
	almaMirror = "https://repo.almalinux.org/almalinux/"
)

var Alma = OS{
	Name:           "alma",
	PrettyName:     "AlmaLinux",
	Homepage:       "https://almalinux.org/",
	Description:    "Community owned and governed, forever-free enterprise Linux distribution, focused on long-term stability, providing a robust production-grade platform. AlmaLinux OS is binary compatible with RHEL®.",
	ConfigFunction: createAlmaConfigs,
}

func createAlmaConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(almaMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`^AlmaLinux-[\d\.]+-latest-[^-]+-([^-]+)\.iso$`)

	releases := head.NameSortedSubDirs(utils.SemverCompare)
	releases = releases[max(len(releases)-4, 0):]

	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			architectures, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			if id, ok := architectures.SubDirs["isos"]; ok {
				architectures, err = id.Fetch()
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					return
				}
			}
			for _, arch := range three_architectures {
				if d, ok := architectures.SubDirs[string(arch)]; ok {
					contents, err := d.Fetch()
					if err != nil {
						errs <- Failure{Release: release, Arch: arch, Error: err}
						return
					}

					checksums := make(map[string]string)
					if f, ok := contents.Files["CHECKSUM"]; ok {
						checksums, err = cs.Build(cs.Sha256Regex, f.URL)
						if err != nil {
							csErrs <- Failure{Release: release, Arch: arch, Error: err}
							return
						}
					}

					for f, match := range contents.FileMatches(isoRe) {
						checksum := checksums[f.Name]
						ch <- Config{
							Release: release,
							Edition: match[1],
							Arch:    arch,
							ISO: []Source{
								webSource(f.URL.String(), checksum, "", f.Name),
							},
						}
					}
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
