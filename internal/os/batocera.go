package os

import (
	"log"
	"regexp"
	"slices"
	"strconv"
)

const BatoceraMirror = "https://mirrors.o2switch.fr/batocera/x86_64/stable/"

type Batocera struct{}

func (Batocera) Data() OSData {
	return OSData{
		Name:        "batocera",
		PrettyName:  "Batocera",
		Homepage:    "https://batocera.org/",
		Description: "Retro-gaming distribution with the aim of turning any computer/nano computer into a gaming console during a game or permanently.",
	}
}

func (Batocera) CreateConfigs() ([]Config, error) {
	releases, err := getBatoceraReleases()
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`<a href="(batocera-x86_64.*?.img.gz)`)
	ch, errs, wg := getChannels()

	for i := len(releases) - 1; i >= len(releases)-3 && i >= 0; i-- {
		release := strconv.Itoa(releases[i])
		url := BatoceraMirror + release + "/"
		wg.Add(1)
		go func() {
			defer wg.Done()
			page, err := capturePage(url)
			if err != nil {
				errs <- err
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
					webSource(img, "", Gz, ""),
				},
			}
		}()
	}

	return waitForConfigs(ch, errs, &wg), nil
}

func getBatoceraReleases() ([]int, error) {
	page, err := capturePage(BatoceraMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`<a href="([0-9]{2})/"`)
	matches := releaseRe.FindAllStringSubmatch(page, -1)

	releases := make([]int, 0, len(matches))
	for _, match := range matches {
		release, err := strconv.Atoi(match[1])
		if err != nil {
			log.Println(err)
		}
		releases = append(releases, release)
	}
	slices.Sort(releases)
	return releases, nil
}
