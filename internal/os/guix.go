package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	guixDlMirror   = "https://ftpmirror.gnu.org/gnu/guix/"
	guixDataMirror = "https://mirrors.ibiblio.org/gnu/guix/"
	guixVmImageRe  = `href="(guix-system-vm-image-([\d\.]+).x86_64-linux.qcow2)"`
	guixIsoRe      = `href="(guix-system-install-([\d\.]+).x86_64-linux.iso)"`
)

var Guix = OS{
	Name:           "guix",
	PrettyName:     "Guix",
	Homepage:       "https://guix.gnu.org/",
	Description:    "Distribution of the GNU operating system developed by the GNU Project—which respects the freedom of computer users.",
	ConfigFunction: createGuixConfigs,
}

func createGuixConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(guixDataMirror)
	if err != nil {
		return nil, err
	}
	vmImageRe := regexp.MustCompile(guixVmImageRe)
	isoRe := regexp.MustCompile(guixIsoRe)

	configs := make([]Config, 0)
	for _, match := range vmImageRe.FindAllStringSubmatch(page, -1) {
		url := guixDlMirror + match[1]
		configs = append(configs, Config{
			Release: match[2],
			Edition: "vm-image",
			DiskImages: []Disk{
				{
					Source: urlSource(url),
				},
			},
		})
	}

	for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
		url := guixDlMirror + match[1]
		configs = append(configs, Config{
			Release: match[2],
			Edition: "install-iso",
			ISO: []Source{
				urlSource(url),
			},
		})
	}

	return configs, nil
}
