package utils

import (
	"log"
	"sync"
)

func SpawnDistros(distros ...Distro) []OSData {
	ch := make(chan OSData)
	errs := make(chan error)
	var wg sync.WaitGroup
	for _, distro := range distros {
		os := distro.Data()
		wg.Add(1)
		go func() {
			defer wg.Done()
			configs, err := distro.CreateConfigs()
			if err != nil {
				errs <- err
				return
			}
			os.Releases = configs
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
	return data
}