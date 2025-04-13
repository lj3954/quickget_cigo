package os

import (
	"errors"
	"fmt"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	kolibriMirror    = "https://builds.kolibrios.org"
	kolibriEditionRe = `href="([a-z]{2}_[A-Z]{2})\/"`
)

var kolibriOS = OS{
	Name:           "kolibrios",
	PrettyName:     "KolibriOS",
	Homepage:       "https://kolibrios.org/",
	Description:    "Tiny yet incredibly powerful and fast operating system.",
	ConfigFunction: createKolibriOSConfigs,
}

func createKolibriOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	editions, numEditions, err := getBasicReleases(kolibriMirror, kolibriEditionRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numEditions)

	release := "latest"
	for edition := range editions {
		mirror := fmt.Sprintf("%s/%s/", kolibriMirror, edition)
		config := Config{
			Release: release,
			Edition: edition,
			GuestOS: quickgetdata.KolibriOS,
		}
		go func() {
			defer wg.Done()
			checksum, iso, err := getKolibriIsoData(mirror + "sha256sums.txt")
			if err != nil {
				csErrs <- Failure{Release: release, Edition: edition, Error: err}
				config.ISO = []Source{
					webSource(mirror+"latest-iso.7z", "", quickgetdata.SevenZip, ""),
				}
			} else {
				config.ISO = []Source{
					webSource(mirror+iso, checksum, quickgetdata.SevenZip, ""),
				}
			}
			ch <- config
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getKolibriIsoData(url string) (string, string, error) {
	page, err := web.CapturePage(url)
	if err != nil {
		return "", "", err
	}
	for _, line := range strings.Split(page, "\n") {
		line := strings.TrimSpace(line)
		if strings.HasSuffix(line, "iso.7z") {
			if data := strings.Fields(line); len(data) == 2 {
				return data[0], data[1], nil
			}
		}
	}
	return "", "", errors.New("No ISO found in checksums")
}
