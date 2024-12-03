package os

import (
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const cbppApi = "https://api.github.com/repos/CBPP/cbpp/releases"

type CBPP struct{}

func (CBPP) Data() OSData {
	return OSData{
		Name:        "crunchbang++",
		PrettyName:  "Crunchbang++",
		Homepage:    "https://crunchbangplusplus.org/",
		Description: "The classic minimal crunchbang feel, now with debian 12 bookworm.",
	}
}

func (CBPP) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	var apiData []GithubAPI
	if err := web.CapturePageToJson(cbppApi, &apiData); err != nil {
		return nil, err
	}
	configs := make([]Config, 0)
	for i := 0; i < len(apiData) && i < 3; i++ {
		data := apiData[i]
		release := data.TagName

		var isoAsset *GithubAsset
		for _, asset := range data.Assets {
			if strings.Contains(asset.Name, "amd64") {
				isoAsset = &asset
				break
			}
		}
		if isoAsset == nil {
			continue
		}

		var checksum string
		for _, line := range strings.Split(data.Body, "\n") {
			if strings.Contains(line, isoAsset.Name) {
				checksum = strings.SplitN(line, " ", 2)[0]
				break
			}
		}

		configs = append(configs, Config{
			Release: release,
			ISO: []Source{
				urlChecksumSource(isoAsset.URL, checksum),
			},
		})
	}
	return configs, nil
}
