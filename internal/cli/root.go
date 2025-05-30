package cli

import (
	"bytes"
	"encoding/json"
	"log"

	system "os"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/quickemu-project/quickget_configs/internal/os"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

func Launch() {
	distros, status := utils.SpawnDistros(os.List...)
	distros = fixList(distros)

	if err := status.Finalize(); err != nil {
		log.Printf("Failed to create status webpage: %s", err)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(distros); err != nil {
		log.Fatalln(err)
	}
	rawJson := buf.Bytes()

	if err := writeData(rawJson, "quickget_data.json", None); err != nil {
		log.Printf("Could not write uncompressed JSON: %s", err)
	}
	if err := writeData(rawJson, "quickget_data.json.gz", Gzip); err != nil {
		log.Printf("Could not write gzip-compressed JSON: %s", err)
	}
	if err := writeData(rawJson, "quickget_data.json.zst", Zstd); err != nil {
		log.Printf("Could not write zstd-compressed JSON: %s", err)
	}

	enc = json.NewEncoder(system.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(distros); err != nil {
		log.Fatalln(err)
	}
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
	_ = compressionType(iota)
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
