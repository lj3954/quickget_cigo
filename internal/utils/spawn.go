package utils

import (
	"errors"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/quickemu-project/quickget_configs/internal/status"
	"github.com/quickemu-project/quickget_configs/internal/web"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

func SpawnDistros(distros ...OS) ([]OSData, *status.Status) {
	ch := make(chan OSData)
	var wg sync.WaitGroup
	status := status.Create(len(distros))
	for _, distro := range distros {
		os := OSData{
			Name:        distro.Name,
			PrettyName:  distro.PrettyName,
			Description: distro.Description,
			Homepage:    distro.Homepage,
		}
		if distro.ConfigFunction == nil {
			status.FailedOS(os, errors.New("Config function is nil"))
			continue
		}

		failures := make(chan Failure)
		csErrs := make(chan Failure)

		var failureSlice, csFailureSlice []Failure
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

		wg.Add(1)
		go func() {
			defer wg.Done()
			configs, err := distro.ConfigFunction(failures, csErrs)

			if err != nil {
				status.FailedOS(os, err)
				return
			}
			configs = web.RemoveInvalidConfigs(configs, failures, csErrs)

			close(failures)
			close(csErrs)

			if len(configs) == 0 {
				for _, failure := range failureSlice {
					log.Printf("Failure: %s", failure)
				}
				status.FailedOS(os, errors.New("no valid configs found"))
				return
			}

			os.Releases = fixConfigs(configs)
			status.AddOS(os, failureSlice, csFailureSlice)
			if len(configs) > 0 {
				ch <- os
			}
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

func fixConfigs(configs []Config) []Config {
	// We want to sort releases in descending order; editions are typically strings and should be sorted lexicographically
	slices.SortFunc(configs, func(a, b Config) int {
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
	for i := range configs {
		config := &configs[i]
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
	return slices.DeleteFunc(configs, func(c Config) bool {
		return c.Arch != qgdata.X86_64 && c.Arch != qgdata.Aarch64 && c.Arch != qgdata.Riscv64
	})
}
