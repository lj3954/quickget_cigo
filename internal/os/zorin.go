package os

import (
	"errors"
	"strconv"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

const (
	zorinMirror        = "https://zrn.co/"
	zorinReleaseMirror = "https://mirrors.edge.kernel.org/zorinos-isos/"
	zorinReleaseRe     = `href="(\d+)\/"`
)

var Zorin = OS{
	Name:           "zorin",
	PrettyName:     "Zorin OS",
	Homepage:       "https://zorin.com/os",
	Description:    "Alternative to Windows and macOS designed to make your computer faster, more powerful, secure, and privacy-respecting.",
	ConfigFunction: createZorinConfigs,
}

func createZorinConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	c := mirror.HttpClient{}
	head, err := c.ReadDir(zorinReleaseMirror)
	if err != nil {
		return nil, err
	}

	zorinEditions := []string{"core", "lite", "education"}

	ch, wg := getChannels()
	for k, releaseDir := range head.SubDirs {
		// Filter out non-numeric "releases" by parsing as a float
		if _, err := strconv.ParseFloat(k, 64); err != nil {
			continue
		}
		wg.Go(func() {
			release := releaseDir.Name

			checksums := make(map[string]string)
			contents, err := releaseDir.Fetch(c)
			// Directory contents are only used for checksums in this case
			if err != nil {
				csErrs <- Failure{Release: release, Error: err}
			} else {
				for k, f := range contents.Files {
					k = strings.ToLower(k)
					if strings.HasSuffix(k, ".txt") && strings.Contains(k, "sum") {
						checksums, err = cs.Build(cs.Whitespace, f.URL)
						if err != nil {
							csErrs <- Failure{Release: release, Error: err}
						} else {
							break
						}
					}
				}
			}

			for _, edition := range zorinEditions {
				url := zorinMirror + release + edition + "64"

				finalUrl, err := web.FinalRedirectUrl(url)
				if err != nil {
					errs <- Failure{Release: release, Error: err}
					continue
				}
				fields := strings.Split(finalUrl, "/")
				if len(fields) == 0 {
					errs <- Failure{Release: release, Error: errors.New("final url has no fields")}
					continue
				}
				filename := fields[len(fields)-1]
				checksum := checksums[filename]

				ch <- Config{
					Release: release,
					Edition: edition,
					ISO: []Source{
						webSource(url, checksum, "", filename),
					},
				}
			}
		})
	}

	return waitForConfigs(ch, wg), nil
}
