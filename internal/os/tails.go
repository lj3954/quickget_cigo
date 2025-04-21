package os

import (
	"errors"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const tailsApi = "https://tails.boum.org/install/v2/Tails/amd64/stable/latest.json"

var Tails = OS{
	Name:           "tails",
	PrettyName:     "Tails",
	Homepage:       "https://tails.net/",
	Description:    "Portable operating system that protects against surveillance and censorship.",
	ConfigFunction: createTailsConfigs,
}

func createTailsConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var data TailsData
	if err := web.CapturePageToJson(tailsApi, &data); err != nil {
		return nil, err
	}

	var configs []Config
	for _, installation := range data.Installations {
		release := installation.Version
		installationPath := findTailsIso(installation.InstallationPaths)
		if len(installationPath.TargetFiles) == 0 {
			errs <- Failure{Release: release, Error: errors.New("List of target files is empty")}
			continue
		}
		var sources []Source
		for _, targetFile := range installationPath.TargetFiles {
			sources = append(sources, urlChecksumSource(targetFile.Url, targetFile.Sha256))
		}

		configs = append(configs, Config{
			Release: release,
			ISO:     sources,
		})
	}
	return configs, nil
}

func findTailsIso(installations []TailsInstallationPath) *TailsInstallationPath {
	for _, path := range installations {
		if path.Type == "iso" {
			return &path
		}
	}
	return nil
}

type TailsData struct {
	Installations []struct {
		InstallationPaths []TailsInstallationPath `json:"installation-paths"`
		Version           string                  `json:"version"`
	} `json:"installations"`
}

type TailsInstallationPath struct {
	TargetFiles []struct {
		Url    string `json:"url"`
		Sha256 string `json:"sha256"`
	} `json:"target-files"`
	Type string `json:"type"`
}
