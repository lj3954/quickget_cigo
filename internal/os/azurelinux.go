package os

import "github.com/quickemu-project/quickget_configs/internal/cs"

var azureLinux = OS{
	Name:           "azurelinux",
	PrettyName:     "Azure Linux",
	Homepage:       "https://github.com/microsoft/azurelinux",
	Description:    "Linux OS for Azure 1P services and edge appliances",
	ConfigFunction: createAzureLinuxConfigs,
}

func createAzureLinuxConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	ch, wg := getChannelsWith(len(x86_64_aarch64))
	for _, arch := range x86_64_aarch64 {
		go func() {
			defer wg.Done()
			urlBase := "https://aka.ms/azurelinux-3.0-" + string(arch)
			url := urlBase + ".iso"

			csUrl := urlBase + "-iso-checksum"
			checksum, err := cs.SingleWhitespace(csUrl)
			if err != nil {
				csErrs <- Failure{Arch: arch, Error: err}
			}

			ch <- Config{
				Arch: arch,
				ISO: []Source{
					urlChecksumSource(url, checksum),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}
