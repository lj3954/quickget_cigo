package os

import (
	"slices"
	"strings"

	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const fedoraJsonUrl = "https://fedoraproject.org/releases.json"

type Fedora struct{}

func (Fedora) Data() OSData {
	return OSData{
		Name:        "fedora",
		PrettyName:  "Fedora",
		Homepage:    "https://fedoraproject.org/",
		Description: "Innovative platform for hardware, clouds, and containers, built with love by you.",
	}
}

func (Fedora) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releaseData, err := getFedoraReleases()
	if err != nil {
		return nil, err
	}

	configs := make([]Config, len(releaseData))
	for i, r := range releaseData {
		source := webSource(r.URL, r.Sha256, r.ArchiveFormat, "")
		config := Config{
			Release: r.Release,
			Edition: r.Edition,
			Arch:    r.Arch,
		}
		if len(r.ArchiveFormat) == 0 {
			config.ISO = []Source{source}
		} else {
			config.DiskImages = []Disk{
				{
					Source: source,
					Format: quickgetdata.Raw,
				},
			}
		}
		configs[i] = config
	}

	return configs, nil
}

func getFedoraReleases() ([]fedoraRelease, error) {
	var releaseData []fedoraRelease
	if err := capturePageToJson(fedoraJsonUrl, &releaseData); err != nil {
		return nil, err
	}

	validFedoraFiletypes := []string{"raw.xz", "iso"}
	blacklistedEditions := []string{"Server", "Cloud_Base"}
	releaseData = slices.DeleteFunc(releaseData, func(r fedoraRelease) bool {
		return isExcludedEdition(r.Edition, blacklistedEditions) ||
			!isValidFiletype(r.URL, validFedoraFiletypes)
	})

	for i, r := range releaseData {
		if strings.HasSuffix(r.URL, ".raw.xz") {
			releaseData[i].Edition += "_preinstalled"
			releaseData[i].ArchiveFormat = quickgetdata.Xz
		}
		releaseData[i].Release = strings.ReplaceAll(r.Release, " ", "_")
	}

	releaseData = slices.CompactFunc(releaseData, func(a, b fedoraRelease) bool {
		return a.Release == b.Release && a.Edition == b.Edition && a.Arch == b.Arch
	})

	return releaseData, nil
}

type fedoraRelease struct {
	Release       string `json:"version"`
	Arch          Arch   `json:"arch"`
	URL           string `json:"link"`
	Edition       string `json:"subvariant"`
	Sha256        string `json:"sha256"`
	ArchiveFormat ArchiveFormat
}

func isValidFiletype(filename string, validFiletypes []string) bool {
	return slices.ContainsFunc(validFiletypes, func(f string) bool {
		return strings.HasSuffix(filename, f)
	})
}
