package cli

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/quickemu-project/quickget_configs/internal/os"
	"github.com/quickemu-project/quickget_configs/internal/utils"
)

func Launch() {
	distros := utils.SpawnDistros(
		os.Alma{},
	)

	json, err := json.MarshalIndent(distros, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(json))
}
