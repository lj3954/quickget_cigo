package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	sparkyLinuxMirror       = "https://sourceforge.net/projects/sparkylinux/files/"
	sparkyLinuxEditionRe    = `title="(\w+)" class="folder`
	sparkyLinuxStableIsoRe  = `"(sparkylinux-(\d{1,3}\.\d+)-x86_64-(\w+)\.iso)"`
	sparkyLinuxRollingIsoRe = `sparkylinux-(\d{4}.\d{2})-x86_64-(\w+)\.iso`
	sparkyLinuxChecksumRe   = `[0-9a-f]{64}`
)

var SparkyLinux = OS{
	Name:           "sparkylinux",
	PrettyName:     "SparkyLinux",
	Homepage:       "https://sparkylinux.org/",
	Description:    "Fast, lightweight and fully customizable operating system which offers several versions for different use cases.",
	ConfigFunction: createSparkyLinuxConfigs,
}

func createSparkyLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	dirs, numDirs, err := getBasicReleases(sparkyLinuxMirror, sparkyLinuxEditionRe, -1)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(numDirs)

	stableIsoRe := regexp.MustCompile(sparkyLinuxStableIsoRe)
	rollingIsoRe := regexp.MustCompile(sparkyLinuxRollingIsoRe)
	checksumRe := regexp.MustCompile(sparkyLinuxChecksumRe)

	for dir := range dirs {
		go func() {
			defer wg.Done()
			mirror := sparkyLinuxMirror + dir + "/"
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Edition: dir, Error: err}
				return
			}

			stableMatches := stableIsoRe.FindAllStringSubmatch(page, 3)
			wg.Add(len(stableMatches))
			for _, match := range stableMatches {
				go func() {
					defer wg.Done()
					release := match[2]
					edition := match[3]
					url := mirror + match[1]
					checksum, err := getSparkyLinuxChecksum(url+".allsums.txt/download", checksumRe)
					url += "/download"
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

			if match := rollingIsoRe.FindStringSubmatch(page); match != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					release := match[1]
					edition := match[2]
					url := mirror + match[0]
					checksum, err := getSparkyLinuxChecksum(url+".allsums.txt/download", checksumRe)
					url += "/download"
					if err != nil {
						csErrs <- Failure{Release: release, Edition: edition, Error: err}
					}
					ch <- Config{
						Release: "rolling",
						Edition: edition,
						ISO: []Source{
							urlChecksumSource(url, checksum),
						},
					}
				}()
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

func getSparkyLinuxChecksum(url string, checksumRe *regexp.Regexp) (string, error) {
	page, err := web.CapturePage(url)
	if err != nil {
		return "", err
	}
	return checksumRe.FindString(page), nil
}
