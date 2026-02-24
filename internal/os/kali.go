package os

import (
	"regexp"
	"time"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
)

const kaliMirror = "https://cdimage.kali.org/"

var Kali = OS{
	Name:           "kali",
	PrettyName:     "Kali Linux",
	Homepage:       "https://www.kali.org/",
	Description:    "The most advanced Penetration Testing Distribution.",
	ConfigFunction: createKaliConfigs,
}

type kaliMatch struct {
	dateModified time.Time
	file         mirror.File
}

func createKaliConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(kaliMirror)
	if err != nil {
		return nil, err
	}

	isoRe := regexp.MustCompile(`kali-linux-\d{4}(?:-|\.)[^-]+-installer-(amd64|arm64).iso`)
	ch, wg := getChannels()
	releases := [...]string{"current", "kali-weekly"}

	for _, release := range releases {
		releaseDir, ok := head.SubDirs[release]
		if !ok {
			continue
		}
		wg.Go(func() {
			contents, err := releaseDir.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}

			checksums := make(map[string]string)
			if f, ok := contents.Files["SHA256SUMS"]; ok {
				checksums, err = cs.Build(cs.Whitespace, f)
				if err != nil {
					csErrs <- Failure{Release: release, Error: err}
				}
			}

			// Filter to the latest ISO for kali weekly
			files := make(map[Arch]kaliMatch)
			for f, match := range contents.FileMatches(isoRe) {
				a, v := NewArch(match[1])
				if !v {
					continue
				}
				if v, ok := files[a]; !ok || f.LastModifiedDate.After(v.dateModified) {
					files[a] = kaliMatch{
						dateModified: f.LastModifiedDate,
						file:         f,
					}
				}
			}

			for arch, m := range files {
				f := m.file
				checksum := checksums[f.Name]
				ch <- Config{
					Release: release,
					Arch:    arch,
					ISO: []Source{
						webSource(f.URL.String(), checksum, "", f.Name),
					},
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
