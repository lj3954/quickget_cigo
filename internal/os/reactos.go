package os

import (
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	reactOsLatestRel = "https://sourceforge.net/projects/reactos/files/latest/download"
)

var reactOS = OS{
	Name:           "reactos",
	PrettyName:     "ReactOS",
	Homepage:       "https://reactos.org/",
	Description:    "Imagine running your favorite Windows applications and drivers in an open-source environment you can trust.",
	ConfigFunction: createReactOSConfigs,
}

func createReactOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	url, err := web.FinalRedirectUrl(reactOsLatestRel)
	if err != nil {
		return nil, err
	}

	return []Config{
		{
			Release: "latest",
			Edition: "standard",
			GuestOS: quickgetdata.ReactOS,
			ISO: []Source{
				webSource(url, "", quickgetdata.Zip, ""),
			},
		},
		{
			Release: "latest",
			Edition: "live",
			GuestOS: quickgetdata.ReactOS,
			ISO: []Source{
				webSource(strings.Replace(url, "iso", "live", 1), "", quickgetdata.Zip, ""),
			},
		},
	}, nil
}
