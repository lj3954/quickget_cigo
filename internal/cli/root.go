package cli

import (
	"encoding/json"
	"fmt"
	"log"

	system "os"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/quickemu-project/quickget_configs/internal/os"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

func Launch() {
	distros, status := utils.SpawnDistros(
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
		os.CachyOS{},
	)
	distros = fixList(distros)

	if err := status.Finalize(); err != nil {
		log.Printf("Failed to create status webpage: %s", err)
	}

	rawJson, err := json.Marshal(distros)
	if err != nil {
		log.Fatalln(err)
	}

	if err := writeData(rawJson, "quickget_data.json", None); err != nil {
		log.Println(err)
	}
	if err := writeData(rawJson, "quickget_data.json.gz", Gzip); err != nil {
		log.Println(err)
	}
	if err := writeData(rawJson, "quickget_data.json.zst", Zstd); err != nil {
		log.Println(err)
	}

	prettyJson, err := json.MarshalIndent(distros, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(prettyJson))
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
		}
	}

	return distros
}

type compressionType int

const (
	_ = iota
	None
	Gzip
	Zstd
)

func writeData(data []byte, filename string, compression compressionType) error {
	file, err := system.Create(filename)
	if err != nil {
		return err
	}
	switch compression {
	case None:
		if _, err := file.Write(data); err != nil {
			return err
		}
	case Gzip:
		enc, err := gzip.NewWriterLevel(file, gzip.BestCompression)
		if err != nil {
			return err
		}
		if _, err := enc.Write(data); err != nil {
			return err
		}
		return enc.Close()
	case Zstd:
		enc, err := zstd.NewWriter(file, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		if err != nil {
			return err
		}
		if _, err := enc.Write(data); err != nil {
			return err
		}
		return enc.Close()
	}
	return nil
}
