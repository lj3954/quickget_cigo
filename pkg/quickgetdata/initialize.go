package quickgetdata

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
