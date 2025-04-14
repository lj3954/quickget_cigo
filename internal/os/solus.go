package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	solusMirror = "https://downloads.getsol.us/isos/"
	solusIsoRe  = `^Solus-([^-]+)-Release-\d{4}-\d{2}-\d{2}\.iso$`
)

var Solus = OS{
	Name:           "solus",
	PrettyName:     "Solus",
	Homepage:       "https://getsol.us/",
	Description:    "Designed for home computing. Every tweak enables us to deliver a cohesive computing experience.",
	ConfigFunction: createSolusConfigs,
}

func createSolusConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var releases []SolusData
	if err := web.CapturePageAcceptingJson(solusMirror, &releases); err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(solusIsoRe)
	ch, wg := getChannelsWith(len(releases))
	for _, release := range releases {
		release := strings.TrimSuffix(release.Name, "/")
		url := solusMirror + release + "/"
		go func() {
			defer wg.Done()
			var isoData []SolusData
			if err := web.CapturePageAcceptingJson(url, &isoData); err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			for _, iso := range isoData {
				isoMatch := isoRe.FindStringSubmatch(iso.Name)
				if len(isoMatch) != 2 {
					continue
				}
				url := url + iso.Name
				checksum, err := cs.SingleWhitespace(url + ".sha256sum")
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
				ch <- Config{
					Release: release,
					Edition: isoMatch[1],
					ISO: []Source{
						urlChecksumSource(url, checksum),
					},
				}
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}

type SolusData struct {
	Name string `json:"name"`
}
