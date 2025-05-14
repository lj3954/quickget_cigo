package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	arcoLinuxMirror    = "https://sourceforge.net/projects/arconetpro/files/"
	arcoLinuxEditionRe = `title="(arco[^-]+)" class="folder ">`
	arcoLinuxIsoRe     = `title="((?:arco|arch)[^-]+-(?:20|v)(\d{2}\.\d{2}\.\d{2})-x86_64\.iso)" class="file ">`
)

var ArcoLinux = OS{
	Name:           "arcolinux",
	PrettyName:     "ArcoLinux",
	Homepage:       "https://arcolinux.com/",
	Description:    "It's all about becoming an expert in Linux.",
	ConfigFunction: createArcoLinuxConfigs,
}

func createArcoLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	editions, numEditions, err := getBasicReleases(arcoLinuxMirror, arcoLinuxEditionRe, -1)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(arcoLinuxIsoRe)
	ch, wg := getChannelsWith(numEditions)

	for edition := range editions {
		go func() {
			defer wg.Done()
			mirror := arcoLinuxMirror + edition + "/"
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Edition: edition, Error: err}
				return
			}

			matches := isoRe.FindAllStringSubmatch(page, 3)
			wg.Add(len(matches))
			for _, match := range matches {
				url := mirror + match[1]
				release := match[2]
				go func() {
					defer wg.Done()
					checksum, err := cs.SingleWhitespace(url + ".md5/download")
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}

					ch <- Config{
						Release: release,
						Edition: edition,
						ISO: []Source{
							urlChecksumSource(url+"/download", checksum),
						},
					}
				}()
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
