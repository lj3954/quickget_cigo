package os

import (
	"errors"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	rebornOsDlPage = "https://downloads.rebornos.org/"
	rebornOsCsRe   = `SHA256.*?([a-f0-9]{64})</div>`
	rebornOsUrlRe  = `href="(https://cdn.soulharsh007.dev/RebornOS-ISO/reborn.*?)"`
)

var RebornOS = OS{
	Name:           "rebornos",
	PrettyName:     "RebornOS",
	Homepage:       "https://rebornos.org/",
	Description:    "Aiming to make Arch Linux as user friendly as possible by providing interface solutions to things you normally have to do in a terminal.",
	ConfigFunction: createRebornOSConfigs,
}

func createRebornOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(rebornOsDlPage)
	if err != nil {
		return nil, err
	}
	csRe := regexp.MustCompile(rebornOsCsRe)
	urlRe := regexp.MustCompile(rebornOsUrlRe)

	release := "latest"
	urlResult := urlRe.FindStringSubmatch(page)
	if urlResult == nil {
		return nil, errors.New("Could not find download URL in HTML")
	}
	url := urlResult[1]

	checksumResult := csRe.FindStringSubmatch(page)
	var checksum string
	if checksumResult == nil {
		csErrs <- Failure{Release: release, Error: errors.New("Could not find checksum from HTML")}
	} else {
		checksum = checksumResult[1]
	}
	return []Config{
		{
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		},
	}, nil
}
