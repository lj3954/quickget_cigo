package utils

import (
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

type completion struct {
	Configs []Config
	Err     error
}

func SpawnDistros(distros ...Distro) ([]OSData, *Status) {
	ch := make(chan OSData)
	errs := make(chan error)
	var wg sync.WaitGroup
	status := createStatus(len(distros))
	for _, distro := range distros {
		os := distro.Data()
		failures := make(chan Failure)
		csErrs := make(chan Failure)
		resultCh := make(chan completion)
		wg.Add(1)
		go func() {
			configs, err := distro.CreateConfigs(failures, csErrs)
			close(failures)
			close(csErrs)
			resultCh <- completion{configs, err}
		}()

		failureSlice := make([]Failure, 0)
		csFailureSlice := make([]Failure, 0)
		go func() {
			for failure := range failures {
				failureSlice = append(failureSlice, failure)
			}
		}()
		go func() {
			for failure := range csErrs {
				csFailureSlice = append(csFailureSlice, failure)
			}
		}()

		go func() {
			defer wg.Done()
			result := <-resultCh
			if result.Err != nil {
				status.failedOS(os, result.Err)
				return
			}
			fixConfigs(&result.Configs)
			status.addOS(os, result.Configs, failureSlice, csFailureSlice)
			os.Releases = result.Configs
			ch <- os
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

	data := make([]OSData, 0, len(distros))
	for os := range ch {
		data = append(data, os)
	}
	slices.SortFunc(data, func(a, b OSData) int {
		return strings.Compare(a.Name, b.Name)
	})

	return data, status
}

func fixConfigs(configs *[]Config) {
	// We want to sort releases in descending order; editions are typically strings and should be sorted lexicographically
	slices.SortFunc(*configs, func(a, b Config) int {
		if aSemver, err := version.NewVersion(a.Release); err == nil {
			if bSemver, err := version.NewVersion(b.Release); err == nil {
				if cmp := bSemver.Compare(aSemver); cmp != 0 {
					return cmp
				}
			}
		}

		if cmp := strings.Compare(b.Release, a.Release); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Edition, b.Edition)
	})
	for i := range *configs {
		config := &(*configs)[i]
		if config.GuestOS == "" {
			config.GuestOS = qgdata.Linux
		}
		if config.Arch == "" || config.Arch == "amd64" {
			config.Arch = qgdata.X86_64
		} else if config.Arch == "arm64" {
			config.Arch = qgdata.Aarch64
		}
		if config.Release == "" {
			config.Release = "latest"
		}
	}
}
