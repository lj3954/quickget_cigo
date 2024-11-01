package cli

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/quickemu-project/quickget_configs/internal/os"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

func Launch() {
	distros := utils.SpawnDistros(
		os.Alma{},
		os.Alpine{},
		os.AntiX{},
		os.Archcraft{},
		os.ArchLinux{},
		os.ArcoLinux{},
		os.ArtixLinux{},
		os.AthenaOS{},
		os.Batocera{},
		os.Bazzite{},
		os.BigLinux{},
		os.BlendOS{},
		os.Bodhi{},
		os.BunsenLabs{},
	)
	distros = fixList(distros)

	json, err := json.MarshalIndent(distros, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(json))
}

func fixList(distros []utils.OSData) []utils.OSData {
	for i, distro := range distros {
		for j := range distro.Releases {
			config := &distros[i].Releases[j]
			// Handle default values
			if config.GuestOS == quickgetdata.Linux {
				config.GuestOS = ""
			}
			if config.Arch == quickgetdata.X86_64 {
				config.Arch = ""
			}
			if config.Release == "" {
				config.Release = "latest"
			}
		}
	}

	return distros
}
