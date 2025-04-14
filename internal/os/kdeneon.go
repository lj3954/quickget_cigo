package os

import (
	"fmt"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const kdeNeonMirror = "https://files.kde.org/neon/images/"

var KDENeon = OS{
	Name:           "kdeneon",
	PrettyName:     "KDE Neon",
	Homepage:       "https://neon.kde.org/",
	Description:    "Latest and greatest of KDE community software packaged on a rock-solid base.",
	ConfigFunction: createKdeNeonConfigs,
}

func createKdeNeonConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases := [...]string{"user", "testing", "unstable", "developer"}
	ch, wg := getChannelsWith(len(releases))
	for _, release := range releases {
		mirror := kdeNeonMirror + release + "/current/"
		isoRelease := release
		if isoRelease == "developer" {
			isoRelease = "unstable-developer"
		}
		isoBase := fmt.Sprintf("neon-%s-current", isoRelease)
		url := mirror + isoBase + ".iso"
		checksumUrl := mirror + isoBase + ".sha256sum"

		go func() {
			defer wg.Done()
			checksum, err := cs.SingleWhitespace(checksumUrl)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}
			ch <- Config{
				Release: release,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
