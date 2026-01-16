package os

import (
	"iter"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

const linuxmintMirror = "https://mirrors.kernel.org/linuxmint/stable/"

var LinuxMint = OS{
	Name:           "linuxmint",
	PrettyName:     "Linux Mint",
	Homepage:       "https://linuxmint.com/",
	Description:    "Designed to work out of the box and comes fully equipped with the apps most people need.",
	ConfigFunction: createLinuxMintConfigs,
}

func createLinuxMintConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.LegacyHttpClient{}
	head, err := c.ReadDir(linuxmintMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(`^linuxmint-\d+(?:\.\d+)?-(\w+)-64bit.iso$`)

	subdirs := head.NameSortedSubDirs(utils.SemverCompare)
	fiveMostRecent := subdirs[max(len(subdirs)-5, 0):]

	ch, wg := getChannels()

	for _, releaseDir := range fiveMostRecent {
		wg.Go(func() {
			configs, err := getLinuxMintReleaseConfigs(releaseDir, c, isoRe, csErrs)
			if err != nil {
				errs <- Failure{Release: releaseDir.Name, Error: err}
				return
			}
			for c := range configs {
				ch <- c
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}

func getLinuxMintReleaseConfigs(dir mirror.SubDirEntry, c mirror.Client, isoRe *regexp.Regexp, csErrs chan<- Failure) (iter.Seq[Config], error) {
	release := dir.Name
	contents, err := dir.Fetch(c)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	for k, f := range contents.Files {
		k = strings.ToLower(k)
		if strings.HasSuffix(k, ".txt") && strings.Contains(k, "sum") {
			checksums, err = cs.Build(cs.Whitespace, f)
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			} else {
				break
			}
		}
	}

	return func(yield func(Config) bool) {
		for f, match := range contents.FileMatches(isoRe) {
			edition := match[1]
			checksum := checksums["*"+f.Name]

			c := Config{
				Release: release,
				Edition: edition,
				ISO: []Source{
					webSource(f.URL.String(), checksum, "", f.Name),
				},
			}
			if !yield(c) {
				return
			}
		}
	}, nil
}
