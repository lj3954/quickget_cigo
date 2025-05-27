package os

import "github.com/quickemu-project/quickget_configs/internal/cs"

const (
	tuxedoMirror      = "https://os.tuxedocomputers.com/"
	tuxedoIsoFilename = "TUXEDO-OS_current.iso"
)

var Tuxedo = OS{
	Name:           "tuxedo-os",
	PrettyName:     "Tuxedo OS",
	Homepage:       "https://tuxedocomputers.com/",
	Description:    "KDE Ubuntu LTS designed to go with their Linux hardware.",
	ConfigFunction: createTuxedoConfigs,
}

func createTuxedoConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	url := tuxedoMirror + tuxedoIsoFilename
	csUrl := tuxedoMirror + "checksums/" + tuxedoIsoFilename + ".sha256"
	cs, err := cs.SingleWhitespace(csUrl)
	if err != nil {
		csErrs <- Failure{Error: err}
	}

	return []Config{
		{
			ISO: []Source{
				urlChecksumSource(url, cs),
			},
		},
	}, nil
}
