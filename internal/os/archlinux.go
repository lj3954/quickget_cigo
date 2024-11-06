package os

import (
	"encoding/json"
)

const (
	ArchLinuxAPI    = "https://archlinux.org/releng/releases/json/"
	ArchLinuxMirror = "https://mirror.rackspace.com/archlinux"
)

type ArchLinux struct{}

func (ArchLinux) Data() OSData {
	return OSData{
		Name:        "archlinux",
		PrettyName:  "Arch Linux",
		Homepage:    "https://archlinux.org/",
		Description: "Lightweight and flexible Linux® distribution that tries to Keep It Simple.",
	}
}

func (ArchLinux) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := capturePage(ArchLinuxAPI)
	if err != nil {
		return nil, err
	}
	var apiData archAPI
	if err := json.Unmarshal([]byte(page), &apiData); err != nil {
		return nil, err
	}

	numConfigs := max(3, len(apiData.Releases))
	configs := make([]Config, numConfigs)
	for i := 0; i < numConfigs; i++ {
		data := apiData.Releases[i]
		release := data.Version
		if release == apiData.LatestVersion {
			release = "latest"
		}
		url := ArchLinuxMirror + data.IsoURL
		configs[i] = Config{
			Release: release,
			ISO: []Source{
				urlChecksumSource(url, data.Sha256Sum),
			},
		}
	}

	return configs, nil
}

type archAPI struct {
	Releases []struct {
		Version   string `json:"version"`
		Sha256Sum string `json:"sha256_sum,omitempty"`
		IsoURL    string `json:"iso_url"`
	} `json:"releases"`
	LatestVersion string `json:"latest_version"`
}
