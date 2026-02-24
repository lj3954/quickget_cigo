package os

import (
	"errors"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	slaxMirror = "https://ftp.fi.muni.cz/pub/linux/slax/"
)

var Slax = OS{
	Name:           "slax",
	PrettyName:     "Slax",
	Homepage:       "https://slax.org/",
	Description:    "Compact, fast, and modern Linux operating system that combines sleek design with modular approach. With the ability to run directly from a USB flash drive without the need for installation, Slax is truly portable and fits easily in your pocket.",
	ConfigFunction: createSlaxConfigs,
}

func createSlaxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(slaxMirror)
	if err != nil {
		return nil, err
	}

	var configs []Config
	release := "latest"

	debianRelease, ok := head.FindSubDir(func(d mirror.SubDirEntry) bool {
		return strings.Contains(d.Name, "debian")
	})
	if ok {
		edition := "debian"
		debianConfig, err := getSlaxConfig(release, edition, debianRelease, csErrs)
		if err != nil {
			errs <- Failure{Release: release, Edition: edition, Error: err}
		} else {
			configs = append(configs, *debianConfig)
		}
	}

	slackwareRelease, ok := head.FindSubDir(func(d mirror.SubDirEntry) bool {
		return strings.Contains(d.Name, "slackware")
	})
	if ok {
		edition := "slackware"
		slackwareConfig, err := getSlaxConfig(release, edition, slackwareRelease, csErrs)
		if err != nil {
			errs <- Failure{Release: release, Edition: edition, Error: err}
		} else {
			configs = append(configs, *slackwareConfig)
		}
	}

	return configs, nil
}

func getSlaxConfig(release, edition string, dir mirror.SubDirEntry, csErrs chan<- Failure) (*Config, error) {
	contents, err := dir.Fetch()
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	if f, ok := contents.Files["md5.txt"]; ok {
		checksums, err = cs.Build(cs.Whitespace, f)
		if err != nil {
			csErrs <- Failure{Release: release, Edition: edition, Error: err}
		}
	}

	for k, f := range contents.Files {
		k = strings.ToLower(k)
		if strings.HasSuffix(k, ".iso") && !strings.Contains(k, "32bit") {
			checksum := checksums[f.Name]
			return &Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					webSource(f.URL.String(), checksum, "", f.Name),
				},
			}, nil
		}
	}
	return nil, errors.New("could not find a matching ISO")
}
