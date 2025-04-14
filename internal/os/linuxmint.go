package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	linuxmintMirror    = "https://mirrors.kernel.org/linuxmint/stable/"
	linuxmintReleaseRe = `href="(\d+(?:\.\d+)?)\/"`
)

var LinuxMint = OS{
	Name:           "linuxmint",
	PrettyName:     "Linux Mint",
	Homepage:       "https://linuxmint.com/",
	Description:    "Designed to work out of the box and comes fully equipped with the apps most people need.",
	ConfigFunction: createLinuxMintConfigs,
}

func createLinuxMintConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, numReleases, err := getReverseReleases(linuxmintMirror, linuxmintReleaseRe, 5)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`href="(linuxmint-\d+(?:\.\d+)?-(\w+)-64bit.iso)"`)
	ch, wg := getChannelsWith(numReleases)

	for release := range releases {
		mirror := linuxmintMirror + release + "/"
		checksumUrl := mirror + "sha256sum.txt"
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			checksums, err := cs.Build(cs.Whitespace{}, checksumUrl)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
			matches := isoRe.FindAllStringSubmatch(page, -1)
			for _, match := range matches {
				iso, edition := match[1], match[2]
				url := mirror + iso
				checksum := checksums["*"+iso]
				ch <- Config{
					Release: release,
					Edition: edition,
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}
