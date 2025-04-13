package os

import (
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	bazziteWorkflow = "https://raw.githubusercontent.com/ublue-os/bazzite/main/.github/workflows/build_iso.yml"
	bazziteMirror   = "https://download.bazzite.gg/"
)

var bazzite = OS{
	Name:           "bazzite",
	PrettyName:     "Bazzite",
	Homepage:       "https://bazzite.gg/",
	Description:    "Container native gaming and a ready-to-game SteamOS like.",
	ConfigFunction: createBazziteConfigs,
}

func createBazziteConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(bazziteWorkflow)
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
		url := bazziteMirror + match[1] + "-stable.iso"
		if isExcludedEdition(edition, excludedEditions) {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			checksum, err := cs.SingleWhitespace(url + "-CHECKSUM")
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

	return waitForConfigs(ch, wg), nil
}

func isExcludedEdition(edition string, excludedEditions []string) bool {
	return slices.ContainsFunc(excludedEditions, func(e string) bool {
		return strings.Contains(edition, e)
	})
}
