package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const biglinuxMirror = "https://iso.biglinux.com.br/"

var BigLinux = OS{
	Name:           "biglinux",
	PrettyName:     "BigLinux",
	Homepage:       "https://www.biglinux.com.br/",
	Description:    "It's the right choice if you want to have an easy and enriching experience with Linux. It has been perfected over more than 19 years, following our motto: 'In search of the perfect system'",
	ConfigFunction: createBigLinuxConfigs,
}

func createBigLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(biglinuxMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`^biglinux_([0-9]{4}(?:-[0-9]{2}){2})_(.*?)\.iso$`)
	ch, wg := getChannels()

	for f, match := range head.FileMatches(isoRe) {
		wg.Go(func() {
			release, edition := match[1], match[2]
			var checksum string
			if cf, ok := head.Files[f.Name+".md5"]; ok {
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
		})
	}

	return waitForConfigs(ch, wg), nil
}
