package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const windowsServerMirror = "https://www.microsoft.com/en-us/evalcenter/download-windows-server"

type WindowsServer struct{}

func (WindowsServer) Data() OSData {
	return OSData{
		Name:        "windows-server",
		PrettyName:  "Windows Server",
		Homepage:    "https://www.microsoft.com/en-us/windows-server/",
		Description: "Platform for building an infrastructure of connected applications, networks, and web services.",
	}
}
func (WindowsServer) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases := [...]string{"2025", "2022", "2019", "2016"}
	isoRe := regexp.MustCompile(`scope="row"> (.*?) <\/th>.*?ISO.*?data-target="(https:.*?)"`)
	ch, wg := getChannelsWith(len(releases))

	for _, release := range releases {
		go func() {
			defer wg.Done()
			mirror := windowsServerMirror + "-" + release + "/"
			page, err := web.CapturePage(mirror)
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			for _, match := range isoRe.FindAllStringSubmatch(page, -1) {
				edition := match[1]
				url := match[2]
				ch <- Config{
					Release: release,
					Edition: edition,
					ISO: []Source{
						urlSource(url),
					},
				}
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
