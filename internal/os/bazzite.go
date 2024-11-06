package os

import (
	"regexp"
	"strings"
)

const (
	BazziteWorkflow = "https://raw.githubusercontent.com/ublue-os/bazzite/main/.github/workflows/build_iso.yml"
	BazziteMirror   = "https://download.bazzite.gg/"
)

type Bazzite struct{}

func (Bazzite) Data() OSData {
	return OSData{
		Name:        "bazzite",
		PrettyName:  "Bazzite",
		Homepage:    "https://bazzite.gg/",
		Description: "Container native gaming and a ready-to-game SteamOS like.",
	}
}

func (Bazzite) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := capturePage(BazziteWorkflow)
	if err != nil {
		return nil, err
	}
	workflowRe := regexp.MustCompile(`- (bazzite-?(.*))`)

	excludedEditions := []string{"nvidia", "ally", "asus"}
	ch, wg := getChannels()
	release := "latest"

	for _, match := range workflowRe.FindAllStringSubmatch(page, -1) {
		edition := match[2]
		if edition == "" {
			edition = "plasma"
		} else if len(edition) <= 4 {
			edition += "-plasma"
		}
		url := BazziteMirror + match[1] + "-stable.iso"
		if isExcludedEdition(edition, excludedEditions) {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			checksum, err := singleWhitespaceChecksum(url + "-CHECKSUM")
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

	return waitForConfigs(ch, &wg), nil
}

func isExcludedEdition(edition string, excludedEditions []string) bool {
	for _, excluded := range excludedEditions {
		if strings.Contains(edition, excluded) {
			return true
		}
	}
	return false
}
