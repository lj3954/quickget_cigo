package os

import "github.com/quickemu-project/quickget_configs/internal/web"

const (
	archlinuxAPI    = "https://archlinux.org/releng/releases/json/"
	archlinuxMirror = "https://mirror.rackspace.com/archlinux"
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

func (ArchLinux) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var apiData archAPI
	if err := web.CapturePageToJson(archlinuxAPI, &apiData); err != nil {
		return nil, err
	}

	numConfigs := min(3, len(apiData.Releases))
	configs := make([]Config, numConfigs)
	for i := range numConfigs {
		data := apiData.Releases[i]
		release := data.Version
		if release == apiData.LatestVersion {
			release = "latest"
		}
		url := archlinuxMirror + data.IsoURL
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
