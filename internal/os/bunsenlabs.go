package os

import (
	"log"
	"maps"
	"regexp"
	"sync"
)

const BunsenLabsMirror = "https://ddl.bunsenlabs.org/ddl/"

type BunsenLabs struct{}

func (BunsenLabs) Data() OSData {
	return OSData{
		Name:        "bunsenlabs",
		PrettyName:  "BunsenLabs",
		Homepage:    "https://www.bunsenlabs.org/",
		Description: "Light-weight and easily customizable Openbox desktop. The project is a community continuation of CrunchBang Linux.",
	}
}

func (BunsenLabs) CreateConfigs() ([]Config, error) {
	page, err := capturePage(BunsenLabsMirror)
	if err != nil {
		return nil, err
	}
	releaseRe := regexp.MustCompile(`href="(([^-]+)-1(:?-[0-9]+)?-amd64.hybrid.iso)"`)
	checksums := getBunsenLabsChecksums(page)

	matches := releaseRe.FindAllStringSubmatch(page, -1)
	configs := make([]Config, len(matches))
	for i, match := range releaseRe.FindAllStringSubmatch(page, -1) {
		checksum := checksums[match[1]]
		url := BunsenLabsMirror + match[1]
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

	for _, match := range checksumRe.FindAllStringSubmatch(page, -1) {
		url := BunsenLabsMirror + match[1]
		wg.Add(1)
		go func() {
			defer wg.Done()
			checksums, err := buildChecksum(Whitespace{}, url)
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

	checksums := map[string]string{}
	for cs := range ch {
		maps.Copy(checksums, cs)
	}
	return checksums
}
