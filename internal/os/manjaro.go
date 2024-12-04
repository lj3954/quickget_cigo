package os

import (
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	manjaroJsonUrl     = "https://gitlab.manjaro.org/web/iso-info/-/raw/master/file-info.json"
	manjaroSwayJsonUrl = "https://mirror.manjaro-sway.download/manjaro-sway/release.json"
)

type Manjaro struct{}

func (Manjaro) Data() OSData {
	return OSData{
		Name:        "manjaro",
		PrettyName:  "Manjaro",
		Homepage:    "https://manjaro.org/",
		Description: "Versatile, free, and open-source Linux operating system designed with a strong focus on safeguarding user privacy and offering extensive control over hardware.",
	}
}

func (Manjaro) CreateConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	ch, wg := getChannels()
	wg.Add(2)
	go func() {
		defer wg.Done()
		var data manjaroData
		if err := web.CapturePageToJson(manjaroJsonUrl, &data); err != nil {
			errs <- Failure{Error: err}
		}
		addManjaroConfigs(data.Official, x86_64, ch, wg, csErrs)
		addManjaroConfigs(data.Community, x86_64, ch, wg, csErrs)
		addManjaroConfigs(data.Arm.Generic, aarch64, ch, wg, csErrs)
	}()

	go func() {
		defer wg.Done()
		addManjaroSwayConfig(ch, errs, csErrs)
	}()
	return waitForConfigs(ch, wg), nil
}

type manjaroSwayData struct {
	URL string `json:"url"`
}

func addManjaroSwayConfig(ch chan Config, errs, csErrs chan<- Failure) {
	release := "standard"
	edition := "sway"
	var data []manjaroSwayData
	if err := web.CapturePageToJson(manjaroSwayJsonUrl, &data); err != nil {
		errs <- Failure{Release: release, Edition: edition, Error: err}
	}
	var url string
	for _, e := range data {
		if strings.HasSuffix(e.URL, ".iso") && !strings.Contains(e.URL, "unstable") {
			url = e.URL
			break
		}
	}
	checksum, err := cs.SingleWhitespace(url + ".sha256")
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
}

func addManjaroConfigs(data map[string]manjaroEntry, arch Arch, ch chan Config, wg *sync.WaitGroup, csErrs chan<- Failure) {
	for edition, entry := range data {
		addManjaroConfig(entry, edition, false, arch, ch, wg, csErrs)
	}
}

func addManjaroConfig(entry manjaroEntry, edition string, minimal bool, arch Arch, ch chan Config, wg *sync.WaitGroup, csErrs chan<- Failure) {
	if entry.Minimal != nil {
		addManjaroConfig(*entry.Minimal, edition, true, arch, ch, wg, csErrs)
	}
	if entry.Image == "" {
		return
	}

	wg.Add(1)
	var release string
	if minimal {
		release = "minimal"
	} else {
		release = "standard"
	}
	go func() {
		defer wg.Done()
		checksum, err := cs.SingleWhitespace(entry.Checksum)
		if err != nil {
			csErrs <- Failure{Release: release, Edition: edition, Arch: arch, Error: err}
		}
		config := Config{
			Release: release,
			Edition: edition,
			Arch:    arch,
		}
		if arch == aarch64 {
			config.DiskImages = []Disk{
				{
					Source: webSource(entry.Image, checksum, quickgetdata.Xz, ""),
					Format: quickgetdata.Raw,
				},
			}
		} else {
			config.ISO = []Source{
				urlChecksumSource(entry.Image, checksum),
			}
		}
		ch <- config
	}()
}

type manjaroData struct {
	Official  map[string]manjaroEntry `json:"official"`
	Community map[string]manjaroEntry `json:"community"`
	Arm       struct {
		Generic map[string]manjaroEntry
	} `json:"arm"`
}

type manjaroEntry struct {
	Image     string        `json:"image"`
	Torrent   string        `json:"torrent"`
	Checksum  string        `json:"checksum"`
	Signature string        `json:"signature"`
	Minimal   *manjaroEntry `json:"minimal"`
}
