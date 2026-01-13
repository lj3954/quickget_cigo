package os

import (
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const lmdeMirror = "https://mirrors.edge.kernel.org/linuxmint/debian/"

var LMDE = OS{
	Name:           "lmde",
	PrettyName:     "Linux Mint Debian Edition",
	Homepage:       "https://linuxmint.com/download_lmde.php",
	Description:    "Aims to be as similar as possible to Linux Mint, but without using Ubuntu. The package base is provided by Debian instead.",
	ConfigFunction: createLmdeConfigs,
}

func createLmdeConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(lmdeMirror)
	if err != nil {
		return nil, err
	}
	checksums := make(map[string]string)
	for k, f := range head.Files {
		k = strings.ToLower(k)
		if strings.HasSuffix(k, "txt") && strings.Contains(k, "sum") {
			checksums, err = cs.Build(cs.Whitespace, f.URL)
			if err != nil {
				csErrs <- Failure{Error: err}
			} else {
				break
			}
		}
	}

	isoRe := regexp.MustCompile(`^lmde-(\d+(?:\.\d+)?)-(\w+)-64bit.iso$`)
	var configs []Config

	for f, match := range head.FileMatches(isoRe) {
		checksum := checksums["*"+f.Name]
		configs = append(configs, Config{
			Release: match[1],
			Edition: match[2],
			ISO: []Source{
				webSource(f.URL, checksum, "", f.Name),
			},
		})
	}

	return configs, nil
}
