package os

import "strings"

const (
	zorinMirror        = "https://zrn.co/"
	zorinReleaseMirror = "https://mirrors.edge.kernel.org/zorinos-isos/"
	zorinReleaseRe     = `href="(\d+)\/"`
)

var Zorin = OS{
	Name:           "zorin",
	PrettyName:     "Zorin OS",
	Homepage:       "https://zorin.com/os",
	Description:    "Alternative to Windows and macOS designed to make your computer faster, more powerful, secure, and privacy-respecting.",
	ConfigFunction: createZorinConfigs,
}

func createZorinConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, _, err := getBasicReleases(zorinReleaseMirror, zorinReleaseRe, -1)
	if err != nil {
		return nil, err
	}

	zorinEditions := []string{"core64", "lite64", "education64"}

	var configs []Config
	for release := range releases {
		for _, edition := range zorinEditions {
			url := zorinMirror + release + edition
			configs = append(configs, Config{
				Release: release,
				Edition: strings.TrimSuffix(edition, "64"),
				ISO: []Source{
					urlSource(url),
				},
			})
		}
	}

	return configs, nil
}
