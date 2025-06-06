package os

import (
	"fmt"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	slaxMirror = "https://ftp.fi.muni.cz/pub/linux/slax/"
	slaxIsoRe  = `href="(slax-64bit-(?:debian|slackware)-[\d\.]+\.iso)"`
)

var Slax = OS{
	Name:           "slax",
	PrettyName:     "Slax",
	Homepage:       "https://slax.org/",
	Description:    "Compact, fast, and modern Linux operating system that combines sleek design with modular approach. With the ability to run directly from a USB flash drive without the need for installation, Slax is truly portable and fits easily in your pocket.",
	ConfigFunction: createSlaxConfigs,
}

func createSlaxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	ch, wg := getChannelsWith(2)
	isoRe := regexp.MustCompile(slaxIsoRe)
	release := "latest"

	go func() {
		defer wg.Done()
		edition := "slackware"
		url := slaxMirror + "Slax-Slackware-15.x/"

		page, err := web.CapturePage(url)
		if err != nil {
			errs <- Failure{Release: release, Edition: edition, Error: err}
			return
		}

		isoMatch := isoRe.FindStringSubmatch(page)
		if len(isoMatch) != 2 {
			errs <- Failure{Release: release, Edition: edition, Error: fmt.Errorf("No iso found for %s", edition)}
			return
		}
		iso := isoMatch[1]

		checksums, err := cs.Build(cs.Whitespace, url+"md5.txt")
		if err != nil {
			csErrs <- Failure{Release: release, Edition: edition, Error: err}
		}
		checksum := checksums[iso]

		ch <- Config{
			Release: release,
			Edition: edition,
			ISO: []Source{
				urlChecksumSource(url+iso, checksum),
			},
		}
	}()

	go func() {
		defer wg.Done()
		edition := "debian"
		url := slaxMirror + "Slax-Debian-12.x/"

		page, err := web.CapturePage(url)
		if err != nil {
			errs <- Failure{Release: release, Edition: edition, Error: err}
			return
		}

		isoMatch := isoRe.FindStringSubmatch(page)
		if len(isoMatch) != 2 {
			errs <- Failure{Release: release, Edition: edition, Error: fmt.Errorf("No iso found for %s", edition)}
			return
		}
		iso := isoMatch[1]

		checksums, err := cs.Build(cs.Whitespace, url+"md5.txt")
		if err != nil {
			csErrs <- Failure{Release: release, Edition: edition, Error: err}
		}
		checksum := checksums[iso]

		ch <- Config{
			Release: release,
			Edition: edition,
			ISO: []Source{
				urlChecksumSource(url+iso, checksum),
			},
		}
	}()
	return waitForConfigs(ch, wg), nil
}
