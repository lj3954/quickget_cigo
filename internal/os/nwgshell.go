package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
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
	page, err := web.CapturePage(nwgshellMirror)
	if err != nil {
		return nil, err
	}
	checksums, err := cs.Build(cs.Whitespace, nwgshellMirror+"sha256sums.txt/download")
	if err != nil {
		csErrs <- Failure{Error: err}
	}

	isoRe := regexp.MustCompile(`"name":"(nwg-live-(\d{4}.\d{2}.\d{2})-[^\.]+\.iso)"`)
	matches := isoRe.FindAllStringSubmatch(page, -1)

	configs := make([]Config, len(matches))
	for i, match := range isoRe.FindAllStringSubmatch(page, -1) {
		iso, release := match[1], match[2]
		checksum := checksums[iso]
		url := nwgshellMirror + iso + "/download"

		configs[i] = Config{
			Release: release,
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		}
	}

	return configs, nil
}
