package os

import (
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	slintMirror = "https://slackware.uk/slint/x86_64/"
)

var Slint = OS{
	Name:           "slint",
	PrettyName:     "Slint",
	Homepage:       "https://slint.fr/",
	Description:    "Slint is an easy-to-use, versatile, blind-friendly Linux distribution for 64-bit computers. Slint is based on Slackware and borrows tools from Salix. Maintainer: Didier Spaier.",
	ConfigFunction: createSlintConfigs,
}

func createSlintConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(slintMirror)
	if err != nil {
		return nil, err
	}

	ch, wg := getChannels()

	for release, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch(c)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			if id, e := contents.SubDirs["iso"]; e {
				contents, err = id.Fetch(c)
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					return
				}
			}

			for k, f := range contents.Files {
				if strings.HasSuffix(k, ".iso") {
					var checksum string
					if cf, e := contents.Files[f.Name+".sha256"]; e {
						checksum, err = cs.SingleWhitespace(cf.URL)
						if err != nil {
							csErrs <- Failure{Release: release, Error: err}
						}
					}

					ch <- Config{
						Release: release,
						ISO: []Source{
							webSource(f.URL, checksum, "", f.Name),
						},
					}
					break
				}
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}
