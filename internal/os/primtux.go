package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	primtuxMirror = "https://sourceforge.net/projects/primtux/files/Distribution/"
	primtuxIsoRe  = `"name":"Primtux(\d+)-amd64.iso".*?download_url":"(.*?)"`
)

var primtux = OS{
	Name:           "primtux",
	PrettyName:     "PrimTux",
	Homepage:       "https://primtux.fr/",
	Description:    "A complete and customizable GNU/Linux operating system intended for primary school students and suitable even for older hardware.",
	ConfigFunction: createPrimtuxConfigs,
}

func createPrimtuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(primtuxMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(primtuxIsoRe)
	matches := isoRe.FindAllStringSubmatch(page, -1)

	ch, wg := getChannelsWith(len(matches))
	for _, match := range matches {
		go func() {
			defer wg.Done()
			release := match[1]
			url := match[2]
			checksumUrl := strings.Replace(url, "iso", "md5", 1)
			checksums, err := cs.SingleWhitespace(checksumUrl)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			ch <- Config{
				Release: release,
				ISO: []Source{
					urlChecksumSource(url, checksums),
				},
			}
		}()
	}
	return waitForConfigs(ch, wg), nil
}
