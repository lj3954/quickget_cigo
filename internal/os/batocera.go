package os

import (
	"regexp"
	"slices"
	"strconv"

	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
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

func (Batocera) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getBatoceraReleases(3)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`<a href="(batocera-x86_64.*?.img.gz)`)
	ch, wg := getChannelsWith(len(releases))

	for _, release := range releases {
		release := strconv.Itoa(release)
		url := batoceraMirror + release + "/"
		go func() {
			defer wg.Done()
			page, err := capturePage(url)
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
				Release: release,
				IMG: []Source{
					webSource(img, "", quickgetdata.Gz, ""),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getBatoceraReleases(maxReleases int) ([]int, error) {
	releaseStrings, numReleases, err := getBasicReleases(batoceraMirror, batoceraReleaseRe, -1)
	if err != nil {
		return nil, err
	}

	releases := make([]int, 0, numReleases)
	for releaseString := range releaseStrings {
		if release, err := strconv.Atoi(releaseString); err == nil {
			releases = append(releases, release)
		}
	}
	slices.Sort(releases)
	numFinalReleases := min(len(releases), maxReleases)
	return releases[:len(releases)-numFinalReleases], nil
}
