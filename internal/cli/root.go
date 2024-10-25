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
		for j, release := range distro.Releases {
			// Replace default values with empty strings, to omit them from final JSON
			if release.GuestOS == quickgetdata.Linux {
				distros[i].Releases[j].GuestOS = ""
			}
			if release.Arch == quickgetdata.X86_64 {
				distros[i].Releases[j].Arch = ""
			}
		}
	}

	return distros
}
