package os

import (
	"regexp"
	"slices"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const (
	proxmoxVeMirror = "https://enterprise.proxmox.com/iso/"
	proxmoxIsoRe    = `^proxmox-ve_(\d+\.\d+)-\d+\.iso$`
)

var ProxmoxVE = OS{
	Name:           "proxmox-ve",
	PrettyName:     "Proxmox VE",
	Homepage:       "https://www.proxmox.com/en/proxmox-virtual-environment",
	Description:    "Proxmox Virtual Environment is a complete, open-source server management platform for enterprise virtualization.",
	ConfigFunction: createProxmoxVEConfigs,
}

func createProxmoxVEConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(proxmoxVeMirror)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	if f, e := head.Files["SHA256SUMS"]; e {
		checksums, err = cs.Build(cs.Whitespace, f)
		if err != nil {
			csErrs <- Failure{Error: err}
		}
	}

	isoRe := regexp.MustCompile(proxmoxIsoRe)

	files := slices.Collect(head.MatchingFiles(isoRe))
	slices.SortFunc(files, func(a, b mirror.File) int {
		return a.LastModifiedDate.Compare(b.LastModifiedDate)
	})
	files = files[max(len(files)-2, 0):]

	configs := make([]Config, len(files))
	for i, f := range files {
		// MatchingFiles already validated that the file name here matches the pattern.
		match := isoRe.FindStringSubmatch(f.Name)
		release := match[1]
		checksum := checksums[f.Name]
		configs[i] = Config{
			Release: release,
			ISO: []Source{
				webSource(f.URL.String(), checksum, "", f.Name),
			},
		}
	}

	return configs, nil
}
