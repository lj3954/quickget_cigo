package os

import (
	"errors"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	kolibriMirror    = "https://builds.kolibrios.org"
	kolibriEditionRe = `href="([a-z]{2}_[A-Z]{2})\/"`
)

var KolibriOS = OS{
	Name:           "kolibrios",
	PrettyName:     "KolibriOS",
	Homepage:       "https://kolibrios.org/",
	Description:    "Tiny yet incredibly powerful and fast operating system.",
	ConfigFunction: createKolibriOSConfigs,
}

func createKolibriOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(kolibriMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	release := "latest"
	for edition, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Edition: edition, Error: err}
				return
			}

			checksums := make(map[string]string)
			if cf, ok := contents.Files["sha256sums.txt"]; ok {
				checksums, err = cs.Build(cs.Whitespace, cf)
				if err != nil {
					csErrs <- Failure{Release: release, Edition: edition, Error: err}
				}
			}

			filename := "latest-iso.7z"
			var checksum string
			for k, v := range checksums {
				if strings.HasSuffix(k, "iso.7z") {
					filename = k
					checksum = v
					break
				}
			}

			f, ok := contents.Files[filename]
			if !ok {
				errs <- Failure{Release: release, Edition: edition, Error: errors.New("named iso could not be found in mirror")}
				return
			}

			ch <- Config{
				Release: release,
				Edition: edition,
				GuestOS: quickgetdata.KolibriOS,
				ISO: []Source{
					webSource(f.URL.String(), checksum, quickgetdata.SevenZip, f.Name),
				},
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
