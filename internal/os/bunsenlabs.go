package os

import (
	"maps"
	"regexp"
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const bunsenLabsMirror = "https://ddl.bunsenlabs.org/ddl/"

var BunsenLabs = OS{
	Name:           "bunsenlabs",
	PrettyName:     "BunsenLabs",
	Homepage:       "https://www.bunsenlabs.org/",
	Description:    "Light-weight and easily customizable Openbox desktop. The project is a community continuation of CrunchBang Linux.",
	ConfigFunction: createBunsenLabsConfigs,
}

func createBunsenLabsConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(bunsenLabsMirror)
	if err != nil {
		return nil, err
	}
	isoRe := regexp.MustCompile(`^([^-]+)-1(:?-[0-9]+)?-amd64.hybrid.iso$`)

	checksums := make(map[string]string)
	for k, f := range head.Files {
		if strings.HasSuffix(k, "txt") && strings.Contains(k, "sum") {
			partialChecksums, err := cs.Build(cs.Whitespace, f)
			if err != nil {
				csErrs <- Failure{Error: err}
			} else {
				maps.Copy(checksums, partialChecksums)
			}
		}
	}

	var configs []Config
	for f, match := range head.FileMatches(isoRe) {
		checksum := checksums[f.Name]
		configs = append(configs, Config{
			Release: match[1],
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		})
	}

	return configs, nil
}

func getBunsenLabsChecksums(page string, csErrs chan<- Failure) map[string]string {
	checksumRe := regexp.MustCompile(`href="(.*?.sha256.txt)"`)
	ch := make(chan map[string]string)
	var wg sync.WaitGroup

	matches := checksumRe.FindAllStringSubmatch(page, -1)
	for _, match := range matches {
		url := bunsenLabsMirror + match[1]
		wg.Go(func() {
			checksums, err := cs.Build(cs.Whitespace, url)
			if err != nil {
				csErrs <- Failure{Error: err}
			} else {
				ch <- checksums
			}
		})
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	checksums := make(map[string]string)
	for cs := range ch {
		maps.Copy(checksums, cs)
	}
	return checksums
}
