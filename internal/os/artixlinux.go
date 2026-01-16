package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const artixMirror = "https://mirrors.ocf.berkeley.edu/artix-iso/"

var ArtixLinux = OS{
	Name:           "artixlinux",
	PrettyName:     "Artix Linux",
	Homepage:       "https://artixlinux.org/",
	Description:    "The Art of Linux. Simple. Fast. Systemd-free.",
	ConfigFunction: createArtixLinuxConfigs,
}

func createArtixLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(artixMirror)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	if f, e := head.Files["sha256sums"]; e {
		checksums, err = cs.Build(cs.Whitespace, f)
		if err != nil {
			csErrs <- Failure{Error: err}
		}
	}
	isoRe := regexp.MustCompile(`^artix-(.*?)-([^-]+-[0-9]+)-x86_64.iso$`)

	var configs []Config
	for f, match := range head.FileMatches(isoRe) {
		checksum := checksums[f.Name]
		configs = append(configs, Config{
			Release: match[2],
			Edition: match[1],
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		})
	}
	return configs, nil
}
