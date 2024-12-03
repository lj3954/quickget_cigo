package os

import (
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const biglinuxMirror = "https://iso.biglinux.com.br/"

type BigLinux struct{}

func (BigLinux) Data() OSData {
	return OSData{
		Name:        "biglinux",
		PrettyName:  "BigLinux",
		Homepage:    "https://www.biglinux.com.br/",
		Description: "It's the right choice if you want to have an easy and enriching experience with Linux. It has been perfected over more than 19 years, following our motto: 'In search of the perfect system'",
	}
}

func (BigLinux) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(biglinuxMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`<a href="(biglinux_([0-9]{4}(?:-[0-9]{2}){2})_(.*?).iso)"`)
	matches := isoRe.FindAllStringSubmatch(page, -1)
	slices.SortFunc(matches, func(a, b []string) int {
		return strings.Compare(b[2], a[2])
	})
	ch, wg := getChannelsWith(len(matches))

	for _, match := range matches {
		url := biglinuxMirror + match[1]
		go func() {
			release, edition := match[2], match[3]
			defer wg.Done()
			checksum, err := cs.SingleWhitespace(url + ".md5")
			if err != nil {
				csErrs <- Failure{Release: release, Edition: edition, Error: err}
			}
			ch <- Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
