package os

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const (
	easyosMirror        = "https://distro.ibiblio.org/easyos/amd64/releases/"
	easyosReleaseNameRe = `href="([a-z]+/)"`
)

type EasyOS struct{}

func (EasyOS) Data() OSData {
	return OSData{
		Name:        "easyos",
		PrettyName:  "EasyOS",
		Homepage:    "https://easyos.org/",
		Description: "Experimental distribution designed from scratch to support containers.",
	}
}

func (EasyOS) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	ch, wg := getChannels()
	releases, err := getEasyOSReleases(5, wg, errs)
	if err != nil {
		return nil, err
	}
	wg.Add(len(releases))

	imgRe := regexp.MustCompile(`href="(easy-[0-9.]+-amd64.img(.gz)?)"`)
	for _, relMirror := range releases {
		release, mirror := relMirror.release, relMirror.mirror
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			checksum, err := cs.SingleWhitespace(mirror + "md5sum.txt")
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			}

			imgMatch := imgRe.FindStringSubmatch(page)
			if imgMatch == nil {
				errs <- Failure{Release: release, Error: errors.New("No image found")}
				return
			}
			url := mirror + imgMatch[1]
			var archiveFormat quickgetdata.ArchiveFormat
			if imgMatch[2] != "" {
				archiveFormat = quickgetdata.Gz
			}
			ch <- Config{
				Release: release,
				DiskImages: []Disk{
					{
						Source: webSource(url, checksum, archiveFormat, ""),
						Format: quickgetdata.Raw,
					},
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getEasyOSReleases(maxReleases int, wg *sync.WaitGroup, errs chan<- Failure) ([]relMirror, error) {
	releaseNames, numReleases, err := getBasicReleases(easyosMirror, easyosReleaseNameRe, -1)
	if err != nil {
		return nil, err
	}
	wg.Add(numReleases)
	subdirectoryRe := regexp.MustCompile(`href="([0-9]{4}/)"`)
	releaseRe := regexp.MustCompile(`href="([0-9](?:\.[0-9]+)+)/"`)

	ch := make(chan relMirror)
	relMirrors := make([]relMirror, 0)
	go func() {
		wg.Wait()
		close(ch)
	}()

	for releaseName := range releaseNames {
		mirror := easyosMirror + releaseName
		go func() {
			defer wg.Done()
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Error: err}
				return
			}
			matches := subdirectoryRe.FindAllStringSubmatch(page, -1)
			wg.Add(len(matches))
			for _, match := range matches {
				mirror := mirror + match[1]
				go func() {
					defer wg.Done()
					page, err := web.CapturePage(mirror)
					if err != nil {
						errs <- Failure{Error: err}
						return
					}
					for _, match := range releaseRe.FindAllStringSubmatch(page, -1) {
						ch <- relMirror{
							release: match[1],
							mirror:  mirror + match[1] + "/",
						}
					}
				}()
			}
		}()
	}

	for relMirror := range ch {
		relMirrors = append(relMirrors, relMirror)
	}

	releases := sortEasyOSReleases(relMirrors)
	return releases[:min(len(releases), maxReleases)], nil
}

func sortEasyOSReleases(releases []relMirror) []relMirror {
	slices.SortFunc(releases, func(a, b relMirror) int {
		if aVer, err := version.NewVersion(a.release); err == nil {
			if bVer, err := version.NewVersion(b.release); err == nil {
				if cmp := bVer.Compare(aVer); cmp != 0 {
					return cmp
				}
			}
		}
		return 0
	})

	return slices.CompactFunc(releases, func(a, b relMirror) bool {
		aParts := strings.SplitN(a.release, ".", 3)
		bParts := strings.SplitN(b.release, ".", 3)
		return aParts[0] == bParts[0] && aParts[1] == bParts[1]
	})

}

type relMirror struct {
	release string
	mirror  string
}
