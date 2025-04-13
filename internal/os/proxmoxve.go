package os

import (
	"regexp"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	proxmoxVeMirror = "https://enterprise.proxmox.com/iso/"
	proxmoxIsoRe    = `href="(proxmox-ve_(\d+\.\d+)-\d+\.iso)"`
)

var proxmoxVE = OS{
	Name:           "proxmox-ve",
	PrettyName:     "Proxmox VE",
	Homepage:       "https://www.proxmox.com/en/proxmox-virtual-environment",
	Description:    "Proxmox Virtual Environment is a complete, open-source server management platform for enterprise virtualization.",
	ConfigFunction: createProxmoxVEConfigs,
}

func createProxmoxVEConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	page, err := web.CapturePage(proxmoxVeMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(proxmoxIsoRe)
	matches := isoRe.FindAllStringSubmatch(page, -1)

	checksums, err := cs.Build(cs.Whitespace{}, proxmoxVeMirror+"SHA256SUMS")
	if err != nil {
		csErrs <- Failure{Error: err}
	}

	configs := make([]Config, len(matches))
	for i, match := range matches {
		iso := match[1]
		release := match[2]
		checksum := checksums[iso]
		url := proxmoxVeMirror + iso
		configs[i] = Config{
			Release: release,
			ISO: []Source{
				urlChecksumSource(url, checksum),
			},
		}
	}

	return configs, nil
}
