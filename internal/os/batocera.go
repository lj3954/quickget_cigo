package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	batoceraMirror    = "https://mirrors.o2switch.fr/batocera/x86_64/stable/"
	batoceraReleaseRe = `<a href="([0-9]{2})/"`
)

type Batocera struct{}

func (Batocera) Data() OSData {
	return OSData{
		Name:        "batocera",
		PrettyName:  "Batocera",
		Homepage:    "https://batocera.org/",
		Description: "Retro-gaming distribution with the aim of turning any computer/nano computer into a gaming console during a game or permanently.",
	}
}

func (Batocera) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases, err := getSortedReleasesFunc(batoceraMirror, batoceraReleaseRe, 3, integerCompare)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`<a href="(batocera-x86_64.*?.img.gz)`)
	ch, wg := getChannelsWith(len(releases))

	for _, release := range releases {
		url := batoceraMirror + release + "/"
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(url)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			match := isoRe.FindStringSubmatch(page)
			if match == nil {
				return
			}

			img := url + match[1]
			ch <- Config{
				GuestOS: quickgetdata.Batocera,
				Release: release,
				IMG: []Source{
					webSource(img, "", quickgetdata.Gz, ""),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
