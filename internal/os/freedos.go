package os

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const (
	freedosMirror    = "https://www.ibiblio.org/pub/micro/pc-stuff/freedos/files/distributions/"
	freedosReleaseRe = `href="(\d+\.\d+)/"`
)

var FreeDOS = OS{
	Name:           "freedos",
	PrettyName:     "FreeDOS",
	Homepage:       "https://www.freedos.org/",
	Description:    "DOS-compatible operating system that you can use to play classic DOS games, run legacy business software, or develop embedded systems.",
	ConfigFunction: createFreeDOSConfigs,
}

func createFreeDOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(freedosMirror)
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()
	isoRe := regexp.MustCompile(`^FD\d+-?(.*?CD)\.(iso|zip)$`)

	for release, d := range head.SubDirs {
		wg.Go(func() {
			contents, err := d.Fetch()
			if err != nil {
				errs <- Failure{Release: release, Error: err}
				return
			}
			// FreeDOS releases prior to 1.4 have an "official" subdirectory which must be used.
			// With 1.4, the main directory for the release is used. Handle both cases
			if od, ok := contents.SubDirs["official"]; ok {
				contents, err = od.Fetch()
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					return
				}
			}

			checksums := make(map[string]string)
			for k, f := range contents.Files {
				if k == "verify.txt" {
					contents, err := web.CapturePage(f.URL)
					if err != nil {
						csErrs <- Failure{Release: release, Error: err}
					}
					lines := strings.Split(contents, "\n")
					start, end := slices.Index(lines, "sha256sum:"), slices.Index(lines, "sha512sum:")

					checksums = cs.Whitespace.BuildWithData(strings.Join(lines[start:end], "\n"))
				} else if strings.HasSuffix(k, ".sha") {
					checksums, err = cs.Build(cs.Whitespace, f)
					if err != nil {
						csErrs <- Failure{Release: release, Error: err}
					}
				} else {
					continue
				}
				break
			}

			for f, match := range contents.FileMatches(isoRe) {
				checksum := checksums[f.Name]

				var archiveFormat ArchiveFormat
				if match[2] == "zip" {
					archiveFormat = quickgetdata.Zip
				}

				ch <- Config{
					GuestOS: quickgetdata.FreeDOS,
					Release: release,
					Edition: match[1],
					ISO: []Source{
						webSource(f.URL.String(), checksum, archiveFormat, f.Name),
					},
				}
			}
		})
	}
	return waitForConfigs(ch, wg), nil
}

func getFreeDOSChecksums(url, page string, checksumRe *regexp.Regexp) (map[string]string, error) {
	csUrlMatch := checksumRe.FindString(page)
	if csUrlMatch == "" {
		return nil, errors.New("Could not find Checksum URL")
	}
	return cs.Build(cs.Whitespace, url+csUrlMatch)
}
