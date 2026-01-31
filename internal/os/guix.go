package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	guixDlHost     = "ftpmirror.gnu.org"
	guixDataMirror = "https://mirrors.ibiblio.org/gnu/guix/"
	guixVmImageRe  = `^guix-system-vm-image-([\d\.]+).x86_64-linux.qcow2$`
	guixIsoRe      = `^guix-system-install-([\d\.]+).x86_64-linux.iso$`
)

var Guix = OS{
	Name:           "guix",
	PrettyName:     "Guix",
	Homepage:       "https://guix.gnu.org/",
	Description:    "Distribution of the GNU operating system developed by the GNU Project—which respects the freedom of computer users.",
	ConfigFunction: createGuixConfigs,
}

func createGuixConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(guixDataMirror)
	if err != nil {
		return nil, err
	}
	vmImageRe := regexp.MustCompile(guixVmImageRe)
	isoRe := regexp.MustCompile(guixIsoRe)

	for k := range head.Files {
		head.Files[k].URL.Host = guixDlHost
	}

	configs := make([]Config, 0)
	for f, match := range head.FileMatches(vmImageRe) {
		configs = append(configs, Config{
			Release: match[1],
			Edition: "vm-image",
			DiskImages: []Disk{
				{
					Source: webSource(f.URL.String(), "", "", f.Name),
				},
			},
		})
	}

	for f, match := range head.FileMatches(isoRe) {
		configs = append(configs, Config{
			Release: match[1],
			Edition: "install-iso",
			ISO: []Source{
				webSource(f.URL.String(), "", "", f.Name),
			},
		})
	}

	return configs, nil
}
