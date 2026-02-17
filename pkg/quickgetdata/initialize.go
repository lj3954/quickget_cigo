package quickgetdata

import "strings"

func NewWebSource(url, checksum string, archiveFormat ArchiveFormat, filename string) Source {
	return Source{
		Web: &WebSource{
			URL:           url,
			Checksum:      checksum,
			ArchiveFormat: archiveFormat,
			FileName:      filename,
		},
	}
}

func URLChecksumSource(url, checksum string) Source {
	return NewWebSource(url, checksum, "", "")
}

func URLSource(url string) Source {
	return URLChecksumSource(url, "")
}

func NewDockerSource(url string, privileged bool, sharedDirs []string, filename string) Source {
	return Source{
		Docker: &DockerSource{
			URL:            url,
			Privileged:     privileged,
			SharedDirs:     sharedDirs,
			OutputFilename: filename,
		},
	}
}

func NewArch(input string) (arch Arch, valid bool) {
	switch strings.ToLower(input) {
	case "x86_64":
	case "amd64":
		arch = X86_64
		valid = true
	case "aarch64":
	case "arm64":
		arch = Aarch64
		valid = true
	case "riscv64":
		arch = Riscv64
		valid = true
	}

	return
}
