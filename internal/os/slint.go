package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	slintMirror    = "https://slackware.uk/slint/x86_64/"
	slintReleaseRe = `href="(slint-([\d\.]+)\/)"`
	slintIsoRe     = `href="(slint64-.*?\.iso)"`
)

var slint = OS{
	Name:           "slint",
	PrettyName:     "Slint",
	Homepage:       "https://slint.fr/",
	Description:    "Slint is an easy-to-use, versatile, blind-friendly Linux distribution for 64-bit computers. Slint is based on Slackware and borrows tools from Salix. Maintainer: Didier Spaier.",
	ConfigFunction: createSlintConfigs,
}

func createSlintConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releaseRe := regexp.MustCompile(slintReleaseRe)
	page, err := web.CapturePage(slintMirror)
	if err != nil {
		return nil, err
	}
	matches := releaseRe.FindAllStringSubmatch(page, -1)

	ch, wg := getChannelsWith(len(matches))
	isoRe := regexp.MustCompile(slintIsoRe)

	for _, match := range matches {
		go func() {
			defer wg.Done()
			release := match[2]
			url := slintMirror + match[1] + "iso/"
			page, err := web.CapturePage(url)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			isoMatch := isoRe.FindStringSubmatch(page)
			if len(isoMatch) != 2 {
				errs <- Failure{Release: release, Error: fmt.Errorf("No iso found for %s", release)}
				return
			}
			iso := isoMatch[1]
			url += iso

			checksums, err := cs.Build(cs.Sha256Regex, url+".sha256")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			ch <- Config{
				Release: release,
				ISO: []Source{
					urlChecksumSource(url, checksums[iso]),
				},
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}
