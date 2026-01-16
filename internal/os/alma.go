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
	c := mirror.HttpClient{}
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
			architectures, err := d.Fetch(c)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			if id, e := architectures.SubDirs["isos"]; e {
				architectures, err = id.Fetch(c)
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					return
				}
			}
			for _, arch := range three_architectures {
				if d, e := architectures.SubDirs[string(arch)]; e {
					contents, err := d.Fetch(c)
					if err != nil {
						errs <- Failure{Release: release, Arch: arch, Error: err}
					}

					checksums := make(map[string]string)
					if f, e := contents.Files["CHECKSUM"]; e {
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
								webSource(f.URL, checksum, "", f.Name),
							},
						}
					}
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
