package os

import (
	"errors"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

const (
	deepinMirror = "https://cdimage.deepin.com/releases/"
)

var Deepin = OS{
	Name:           "deepin",
	PrettyName:     "Deepin",
	Homepage:       "https://www.deepin.org/",
	Description:    "Beautiful UI design, intimate human-computer interaction, and friendly community environment make you feel at home.",
	ConfigFunction: createDeepinConfigs,
}

func createDeepinConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(deepinMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	releases := head.NameSortedSubDirs(utils.SemverCompare)
	releases = releases[max(len(releases)-3, 0):]

	for _, d := range releases {
		release := d.Name
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			if len(contents.Files) > 0 {
				config, csErr, err := createDeepinConfig(contents, release, x86_64)
				if err != nil {
					errs <- Failure{Release: release, Error: err}
				} else {
					if csErr != nil {
						csErrs <- Failure{Release: release, Error: err}
					}
					ch <- *config
				}
			}
			for a, d := range contents.SubDirs {
				contents, err := d.Fetch()
				if err != nil {
					errs <- Failure{Release: release, Error: err}
				}
				config, csErr, err := createDeepinConfig(contents, release, Arch(a))
				if err != nil {
					errs <- Failure{Release: release, Error: err}
				} else {
					if csErr != nil {
						csErrs <- Failure{Release: release, Error: err}
					}
					ch <- *config
				}
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}

func createDeepinConfig(dir *mirror.Directory, release string, arch Arch) (config *Config, csErr error, err error) {
	for k, f := range dir.Files {
		if strings.HasSuffix(k, ".iso") {
			var checksum string
			if f, e := dir.Files["SHA256SUMS"]; e {
				checksum, csErr = cs.SingleWhitespace(f)
			}
			config = &Config{
				Release: release,
				Arch:    arch,
				ISO: []Source{
					webSource(f.URL.String(), checksum, "", f.Name),
				},
			}
			return
		}
	}
	return nil, nil, errors.New("could not find ISO file in mirror")
}
