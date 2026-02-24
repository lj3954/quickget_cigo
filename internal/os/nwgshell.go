package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const nwgshellMirror = "https://sourceforge.net/projects/nwg-iso/files/"

var NWGShell = OS{
	Name:           "nwg-shell",
	PrettyName:     "nwg-shell",
	Homepage:       "https://nwg-piotr.github.io/nwg-shell/",
	Description:    "Arch Linux ISO with nwg-shell for sway and Hyprland",
	ConfigFunction: createNwgShellConfigs,
}

func createNwgShellConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.SourceForgeClient{}
	head, err := c.ReadDir(nwgshellMirror)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	if f, ok := head.Files["sha256sums.txt"]; ok {
		checksums, err = cs.Build(cs.Whitespace, f)
		if err != nil {
			csErrs <- Failure{Error: err}
		}
	}

	isoRe := regexp.MustCompile(`^nwg-live-(\d{4}.\d{2}.\d{2})-[^\.]+\.iso$`)

	var configs []Config
	for f, match := range head.FileMatches(isoRe) {
		checksum := checksums[f.Name]

		configs = append(configs, Config{
			Release: match[1],
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		})
	}

	return configs, nil
}
