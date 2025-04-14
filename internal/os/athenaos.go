package os

import (
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const athenaAPI = "https://api.github.com/repos/Athena-OS/athena/releases"

var AthenaOS = OS{
	Name:           "athenaos",
	PrettyName:     "Athena OS",
	Homepage:       "https://athenaos.org/",
	Description:    "Offer a different experience than the most used pentesting distributions by providing only tools that fit with the user needs and improving the access to hacking resources and learning materials.",
	ConfigFunction: createAthenaOSConfigs,
}

func createAthenaOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var apiData []GithubAPI
	if err := web.CapturePageToJson(athenaAPI, &apiData); err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	for i := 0; i < 2 && i < len(apiData); i++ {
		data := apiData[i]
		if data.Assets == nil {
			continue
		}

		release := data.TagName
		if data.Prerelease {
			release += "-pre"
		}

		var isoAsset *GithubAsset
		for _, asset := range data.Assets {
			if strings.HasSuffix(asset.Name, ".iso") {
				isoAsset = &asset
				break
			}
		}
		if isoAsset == nil {
			continue
		}

		checksumName := isoAsset.Name + ".sha256"
		var checksumUrl string
		for _, asset := range data.Assets {
			if asset.Name == checksumName {
				checksumUrl = asset.URL
			}
		}

		if checksumUrl == "" {
			ch <- Config{
				Release: release,
				ISO: []Source{
					urlSource(isoAsset.URL),
				},
			}
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				checksum, err := cs.SingleWhitespace(checksumUrl)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
				ch <- Config{
					Release: release,
					ISO: []Source{
						urlChecksumSource(isoAsset.URL, checksum),
					},
				}
			}()
		}
	}
	return waitForConfigs(ch, wg), nil
}
