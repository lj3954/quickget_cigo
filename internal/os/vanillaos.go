package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	vanillaAPI        = "https://api.github.com/repos/Vanilla-OS/live-iso/releases"
	vanillaRevisionRe = `([\d\.]+)-r[\d\.]+`
)

var VanillaOS = OS{
	Name:           "vanillaos",
	PrettyName:     "Vanilla OS",
	Homepage:       "https://vanillaos.org/",
	Description:    "Designed to be a reliable and productive operating system for your daily work.",
	ConfigFunction: createVanillaOSConfigs,
}

func createVanillaOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var apiData []GithubAPI
	if err := web.CapturePageToJson(vanillaAPI, &apiData); err != nil {
		return nil, err
	}

	revisionRe := regexp.MustCompile(vanillaRevisionRe)
	usedReleases := make(map[string]struct{})
	ch, wg := getChannels()

	for _, entry := range apiData {
		if len(usedReleases) >= 3 {
			break
		}
		release := getVanillaOSRelease(entry.TagName, revisionRe)
		if _, used := usedReleases[release]; used {
			continue
		}
		var isoAsset *GithubAsset
		var checksumUrl string
		for _, asset := range entry.Assets {
			switch {
			case strings.HasSuffix(asset.Name, ".iso"):
				isoAsset = &asset
			case strings.Contains(asset.Name, ".sha256"):
				checksumUrl = asset.URL
			}
		}

		if isoAsset == nil {
			continue
		}
		usedReleases[release] = struct{}{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			var checksum string
			if checksumUrl != "" {
				var err error
				checksum, err = cs.SingleWhitespace(checksumUrl)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
			}
			ch <- Config{
				Release: release,
				ISO: []Source{
					webSource(isoAsset.URL, checksum, "", isoAsset.Name),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getVanillaOSRelease(tag string, revisionRe *regexp.Regexp) string {
	if match := revisionRe.FindStringSubmatch(tag); len(match) == 2 {
		return match[1]
	}
	return tag
}
