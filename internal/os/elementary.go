package os

import (
	"errors"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	elementaryUrl         = "https://elementary.io/"
	elementaryChecksumUrl = "https://elementary.io/docs/installation"
)

type Elementary struct{}

func (Elementary) Data() OSData {
	return OSData{
		Name:        "elementary",
		PrettyName:  "elementary OS",
		Homepage:    "https://elementary.io/",
		Description: "Thoughtful, capable, and ethical replacement for Windows and macOS.",
	}
}

func (Elementary) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := web.CapturePage(elementaryUrl)
	if err != nil {
		return nil, err
	}
	downloadRe := regexp.MustCompile(`download-link http" href="(.*?)">Download`)

	downloadMatch := downloadRe.FindStringSubmatch(page)
	if downloadMatch == nil {
		return nil, errors.New("No download link found in HTML")
	}
	url := "https:" + downloadMatch[1]

	var checksum string
	if csPage, err := web.CapturePage(elementaryChecksumUrl); err != nil {
		csErrs <- Failure{Error: err}
	} else {
		checksumRe := regexp.MustCompile(`"language-bash">([0-9a-f]{64})</code>`)
		csMatch := checksumRe.FindStringSubmatch(csPage)
		if csMatch == nil {
			csErrs <- Failure{Error: errors.New("No checksum found in HTML")}
		} else {
			checksum = csMatch[1]
		}
	}
	return []Config{
		{
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		},
	}, nil
}
