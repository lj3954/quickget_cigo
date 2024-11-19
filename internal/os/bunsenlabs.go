package os

import (
	"log"
	"maps"
	"regexp"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
)

const bunsenLabsMirror = "https://ddl.bunsenlabs.org/ddl/"

type BunsenLabs struct{}

func (BunsenLabs) Data() OSData {
	return OSData{
		Name:        "bunsenlabs",
		PrettyName:  "BunsenLabs",
		Homepage:    "https://www.bunsenlabs.org/",
		Description: "Light-weight and easily customizable Openbox desktop. The project is a community continuation of CrunchBang Linux.",
	}
}

func (BunsenLabs) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	page, err := capturePage(bunsenLabsMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`href="(([^-]+)-1(:?-[0-9]+)?-amd64.hybrid.iso)"`)
	checksums := getBunsenLabsChecksums(page)

	matches := releaseRe.FindAllStringSubmatch(page, -1)
	configs := make([]Config, len(matches))
	for i, match := range releaseRe.FindAllStringSubmatch(page, -1) {
		checksum := checksums[match[1]]
		url := bunsenLabsMirror + match[1]
		configs[i] = Config{
			Release: match[2],
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		}
	}

	return configs, nil
}

func getBunsenLabsChecksums(page string) map[string]string {
	checksumRe := regexp.MustCompile(`href="(.*?.sha256.txt)"`)
	ch := make(chan map[string]string)
	errs := make(chan error)
	var wg sync.WaitGroup

	matches := checksumRe.FindAllStringSubmatch(page, -1)
	wg.Add(len(matches))
	for _, match := range matches {
		url := bunsenLabsMirror + match[1]
		go func() {
			defer wg.Done()
			checksums, err := cs.Build(cs.Whitespace{}, url)
			if err != nil {
				errs <- err
			} else {
				ch <- checksums
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
		close(errs)
	}()
	go func() {
		for err := range errs {
			log.Println(err)
		}
	}()

	checksums := make(map[string]string)
	for cs := range ch {
		maps.Copy(checksums, cs)
	}
	return checksums
}
