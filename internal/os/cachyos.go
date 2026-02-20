package os

import (
	"errors"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

const cachyOSMirror = "https://mirror.cachyos.org/ISO/"

var CachyOS = OS{
	Name:           "cachyos",
	PrettyName:     "CachyOS",
	Homepage:       "https://cachyos.org/",
	Description:    "Designed to deliver lightning-fast speeds and stability, ensuring a smooth and enjoyable computing experience every time you use it.",
	ConfigFunction: createCachyOSConfigs,
}

func createCachyOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(cachyOSMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	for edition, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Error: err}
				return
			}
			releases := contents.NameSortedSubDirs(utils.IntegerCompare)
			for _, d := range slices.Backward(releases) {
				release := "latest"
				contents, err := d.Fetch()
				if err != nil {
					errs <- Failure{Release: d.Name, Edition: edition, Error: err}
					continue
				}

				for k, f := range contents.Files {
					if strings.HasSuffix(k, ".iso") {
						var checksum string
						if cf, e := contents.Files[f.Name+".sha256"]; e {
							checksum, err = cs.SingleWhitespace(cf)
							if err != nil {
								csErrs <- Failure{Release: release, Edition: edition, Error: err}
							}
						}

						ch <- Config{
							Release: release,
							Edition: edition,
							ISO: []Source{
								webSource(f.URL.String(), checksum, "", f.Name),
							},
						}
						// Return as soon as we have a valid config, we only want to keep the latest release for each edition
						return
					}
				}
				errs <- Failure{Release: d.Name, Edition: edition, Error: errors.New("could not find ISO in directory")}
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}
