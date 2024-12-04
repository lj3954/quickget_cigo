package web

import (
	"context"
	"fmt"
	"iter"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/data"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

func RemoveInvalidConfigs(configs []quickgetdata.Config, errs, csErrs chan<- data.Failure) []quickgetdata.Config {
	var wg sync.WaitGroup
	ch := make(chan quickgetdata.Config)
	wg.Add(len(configs))
	for _, config := range configs {
		go func() {
			defer wg.Done()
			if err := validateConfigSources(&config); err != nil {
				errs <- data.Failure{
					Release: config.Release,
					Edition: config.Edition,
					Arch:    config.Arch,
					Error:   err,
				}
				return
			}
			ch <- config
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	var remainingConfigs []quickgetdata.Config
	for config := range ch {
		remainingConfigs = append(remainingConfigs, config)
	}
	return remainingConfigs
}

func validateConfigSources(config *quickgetdata.Config) error {
	sources := concatPointers(
		config.DiskImages,
		config.ISO,
		config.IMG,
		config.FixedISO,
		config.Floppy,
	)
	if err := validateSources(sources); err != nil {
		return err
	}

	return nil
}

func validateSources(sources iter.Seq[*quickgetdata.Source]) error {
	var wg sync.WaitGroup
	errs := make(chan error)
	for source := range sources {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if webSource := source.Web; webSource != nil {
				filename, err := resolveURLFilename(webSource.URL)
				if err != nil {
					errs <- err
				}
				// We want to add filenames wherever possible to simplify the job of quickget.
				// Modifying the URL to the redirect is not desired
				// such redirects could be intended to determine the best available mirror for a location
				if len(webSource.FileName) == 0 {
					webSource.FileName = filename
				}
			} else if dockerSource := source.Docker; dockerSource != nil {
				if _, err := resolveURL(dockerSource.URL); err != nil {
					errs <- err
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(errs)
	}()
	// Handle nil errors potentially entering the channel
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveURL(input string) (*http.Response, error) {
	url, err := url.Parse(input)
	if err != nil {
		return nil, err
	}
	if sem, exists := urlPermits[url.Hostname()]; exists {
		if err := sem.Acquire(context.Background(), 1); err != nil {
			return nil, err
		}
		defer sem.Release(1)
	}
	if err := permits.Acquire(context.Background(), 1); err != nil {
		return nil, err
	}
	defer permits.Release(1)
	resp, err := client.Get(input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	if status == http.StatusTooManyRequests {
		log.Printf("Warning: Got status too many requests for URL %s\n", url)
	} else if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Failed to resolve URL %s: %s", url, resp.Status)
	}
	return resp, nil
}

func resolveURLFilename(url string) (string, error) {
	resp, err := resolveURL(url)
	if err != nil {
		return "", err
	}
	finalUrl := resp.Request.URL
	fields := strings.Split(finalUrl.Path, "/")
	if len(fields) == 0 {
		return "", nil
	}

	filename := fields[len(fields)-1]
	return filename, nil
}

func concatPointers(disks []quickgetdata.Disk, sources ...[]quickgetdata.Source) iter.Seq[*quickgetdata.Source] {
	return func(yield func(*quickgetdata.Source) bool) {
		for i := range sources {
			for j := range sources[i] {
				if !yield(&sources[i][j]) {
					return
				}
			}
		}
		for i := range disks {
			if !yield(&disks[i].Source) {
				return
			}
		}
	}
}
