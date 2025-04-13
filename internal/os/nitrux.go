package os

import (
	"errors"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const nitruxMirror = "https://sourceforge.net/projects/nitruxos/files/Release/"

var nitrux = OS{
	Name:           "nitrux",
	PrettyName:     "Nitrux",
	Homepage:       "https://nxos.org/",
	Description:    "Powered by Debian, KDE Plasma and Frameworks, and AppImages.",
	ConfigFunction: createNitruxConfigs,
}

func createNitruxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(nitruxMirror + "ISO/")
	if err != nil {
		return nil, err
	}
	release := "latest"
	isoRe := regexp.MustCompile(`(nitrux-nx-desktop-plasma-[^-]+-amd64).iso`)
	match := isoRe.FindStringSubmatch(page)
	if len(match) != 2 {
		return nil, errors.New("could not find release in page")
	}
	url := nitruxMirror + "ISO/" + match[0] + "/download"

	isoBase := match[1]
	checksumUrl := nitruxMirror + "SHA512/" + isoBase + ".sha512/download"
	checksum, err := cs.SingleWhitespace(checksumUrl)
	if err != nil {
		csErrs <- Failure{Release: release, Error: err}
	}

	return []Config{
		{
			Release: release,
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		},
	}, nil
}
