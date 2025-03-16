package os

import (
	"errors"
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	windowsRedirectMirror = "https://quickemu-dynamic.lj3954.dev/"
	windowsServerMirror   = "https://www.microsoft.com/en-us/evalcenter/download-windows-server"
)

type Windows struct{}

func (Windows) Data() OSData {
	return OSData{
		Name:        "windows",
		PrettyName:  "Windows",
		Homepage:    "https://www.microsoft.com/en-us/windows/",
		Description: "Whether you’re gaming, studying, running a business, or running a household, Windows helps you get it done.",
	}
}

func (Windows) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	var list []OsListData
	if err := web.CapturePageToJson(windowsRedirectMirror+"list?os=windows", &list); err != nil {
		return nil, err
	}

	var configs []Config
	for _, data := range list {
		if data.Error != "" {
			errs <- Failure{Release: data.Release, Edition: data.Edition, Arch: data.Arch, Error: errors.New(data.Error)}
			continue
		}
		url := windowsRedirectMirror + data.Url[2:]
		configs = append(configs, Config{
			Release: data.Release,
			Edition: data.Edition,
			Arch:    data.Arch,
			ISO: []Source{
				webSource(url, data.Checksum, "", data.Filename),
			},
			SkipValidation: true,
		})
	}

	return configs, nil
}

type OsListData struct {
	Release  string `json:"release"`
	Edition  string `json:"edition"`
	Arch     Arch   `json:"arch"`
	Url      string `json:"url"`
	Filename string `json:"filename,omitempty"`
	Checksum string `json:"checksum,omitempty"`
	Error    string `json:"error,omitempty"`
}

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
