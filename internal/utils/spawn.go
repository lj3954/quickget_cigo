package utils

import (
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/quickemu-project/quickget_configs/internal/web"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

func SpawnDistros(distros ...Distro) ([]OSData, *Status) {
	ch := make(chan OSData)
	var wg sync.WaitGroup
	wg.Add(len(distros))
	status := createStatus(len(distros))
	for _, distro := range distros {
		os := distro.Data()
		failures := make(chan Failure)
		csErrs := make(chan Failure)

		failureSlice := make([]Failure, 0)
		csFailureSlice := make([]Failure, 0)
		go func() {
			for failure := range failures {
				failureSlice = append(failureSlice, failure)
			}
		}()
		go func() {
			for csFailure := range csErrs {
				csFailureSlice = append(csFailureSlice, csFailure)
			}
		}()

		go func() {
			defer wg.Done()
			defer close(failures)
			defer close(csErrs)
			configs, err := distro.CreateConfigs(failures, csErrs)

			if err != nil {
				status.failedOS(os, err)
				return
			}
			configs = web.RemoveInvalidConfigs(configs, failures, csErrs)

			fixConfigs(&configs)
			status.addOS(os, configs, failureSlice, csFailureSlice)
			os.Releases = configs
			ch <- os
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
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
	*configs = slices.DeleteFunc(*configs, func(c Config) bool {
		return c.Arch != qgdata.X86_64 && c.Arch != qgdata.Aarch64 && c.Arch != qgdata.Riscv64
	})
}
