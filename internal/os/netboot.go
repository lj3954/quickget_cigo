package os

import "github.com/quickemu-project/quickget_configs/internal/cs"

const netbootMirror = "https://boot.netboot.xyz/ipxe/"

type Netboot struct{}

func (Netboot) Data() OSData {
	return OSData{
		Name:        "netboot",
		PrettyName:  "Netboot",
		Homepage:    "https://netboot.xyz/",
		Description: "Your favorite operating systems in one place.",
	}
}

func (Netboot) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	checksumUrl := netbootMirror + "netboot.xyz-sha256-checksums.txt"
	checksums, err := cs.Build(cs.Whitespace{}, checksumUrl)
	if err != nil {
		csErrs <- Failure{Error: err}
	}

	return []Config{
		getNetbootConfig(checksums, "netboot.xyz.iso", x86_64),
		getNetbootConfig(checksums, "netboot.xyz-arm64.iso", aarch64),
	}, nil
}

func getNetbootConfig(checksums map[string]string, iso string, arch Arch) Config {
	url := netbootMirror + iso
	checksum := checksums["*"+iso]
	return Config{
		Arch: arch,
		ISO: []Source{
			urlChecksumSource(url, checksum),
		},
	}
}
