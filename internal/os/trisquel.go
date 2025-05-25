package os

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	trisquelMirror = "https://mirrors.ocf.berkeley.edu/trisquel-images/"
	trisquelIsoRe  = `href="(([^_\s]+)_([^_\s]+)_amd64\.iso)"`
)

var Trisquel = OS{
	Name:           "trisquel",
	PrettyName:     "Trisquel",
	Homepage:       "https://trisquel.info/",
	Description:    "Fully free operating system for home users, small enterprises and educational centers.",
	ConfigFunction: createTrisquelConfigs,
}

func createTrisquelConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(trisquelMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(trisquelIsoRe)
	matches := isoRe.FindAllStringSubmatch(page, -1)

	fmt.Println(matches)

	ch, wg := getChannelsWith(len(matches))
	for _, match := range matches {
		go func() {
			defer wg.Done()
			url := trisquelMirror + match[1]
			release := match[3]
			edition := friendlyTrisquelEdition(match[2])

			cs, err := cs.SingleWhitespace(url + ".sha256")
			if err != nil {
				csErrs <- Failure{Release: release, Edition: edition, Error: err}
			}

			ch <- Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					urlChecksumSource(url, cs),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func friendlyTrisquelEdition(edition string) string {
	switch edition {
	case "trisquel":
		return "mate"
	case "trisquel-mini":
		return "lxde"
	case "trisquel-sugar":
		return "sugar"
	case "triskel":
		return "kde"
	default:
		return strings.TrimPrefix(edition, "trisquel-")
	}
}
