package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const endeavourMirror = "https://mirror.alpix.eu/endeavouros/iso/"

var EndeavourOS = OS{
	Name:           "endeavouros",
	PrettyName:     "EndeavourOS",
	Homepage:       "https://endeavouros.com/",
	Description:    "Provides an Arch experience without the hassle of installing it manually for both x86_64 and ARM systems.",
	ConfigFunction: createEndeavourOSConfigs,
}

func createEndeavourOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(endeavourMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`^EndeavourOS_[^\d]+(\d{4}.\d{2}.\d{2}).iso$`)
	ch, wg := getChannels()

	for f, match := range head.FileMatches(isoRe) {
		release := match[1]
		wg.Go(func() {
			cf, ok := head.FindFile(func(f2 mirror.File) bool {
				return strings.HasPrefix(f2.Name, f.Name) && strings.HasSuffix(f2.Name, "sum")
			})
			var checksum string
			if ok {
				checksum, err = cs.SingleWhitespace(cf)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
			}
			ch <- Config{
				Release: release,
				ISO: []Source{
					webSource(f.URL.String(), checksum, "", f.Name),
				},
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
