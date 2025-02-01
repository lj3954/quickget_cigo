package os

import "github.com/quickemu-project/quickget_configs/internal/web"

const rebornOsApi = "https://meta.cdn.soulharsh007.dev/RebornOS-ISO?format=json"

type RebornOS struct{}

func (RebornOS) Data() OSData {
	return OSData{
		Name:        "rebornos",
		PrettyName:  "RebornOS",
		Homepage:    "https://rebornos.org/",
		Description: "Aiming to make Arch Linux as user friendly as possible by providing interface solutions to things you normally have to do in a terminal.",
	}
}

func (RebornOS) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var data rebornOsData
	if err := web.CapturePageToJson(rebornOsApi, &data); err != nil {
		return nil, err
	}
	return []Config{
		{
			ISO: []Source{
				urlChecksumSource(data.URL, data.Checksum),
			},
		},
	}, nil
}

type rebornOsData struct {
	URL      string `json:"url"`
	Checksum string `json:"md5"`
}
