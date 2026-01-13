package os

import (
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	garudaMirror    = "https://iso.builds.garudalinux.org/iso/latest/garuda/"
	garudaEditionRe = `href="([^.]+)\/"`
)

var Garuda = OS{
	Name:           "garuda",
	PrettyName:     "Garuda Linux",
	Homepage:       "https://garudalinux.org/",
	Description:    "Feature rich and easy to use Linux distribution.",
	ConfigFunction: createGarudaConfigs,
}

func createGarudaConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(garudaMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	release := "latest"
	for edition, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch(c)
			if err != nil {
				errs <- Failure{Release: release, Edition: edition, Error: err}
				return
			}

			for k, f := range contents.Files {
				if strings.HasSuffix(k, "iso") {
					var checksum string
					if cf, e := contents.Files[k+".sha256"]; e {
						checksum, err = cs.SingleWhitespace(cf.URL)
						if err != nil {
							csErrs <- Failure{Release: release, Edition: edition, Error: err}
						}
					}

					ch <- Config{
						Release: release,
						Edition: edition,
						ISO: []Source{
							webSource(f.URL, checksum, "", f.Name),
						},
					}
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
