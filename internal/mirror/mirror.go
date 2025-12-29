package mirror

import (
	"maps"
	"slices"
	"time"
)

type Directory struct {
	Name    string
	URL     string
	Files   map[string]File
	SubDirs map[string]SubDirEntry
}

func (d *Directory) ModifiedTimeSortedSubdirs() []SubDirEntry {
	subdirs := slices.Collect(maps.Values(d.SubDirs))
	slices.SortFunc(subdirs, func(a, b SubDirEntry) int {
		return a.LastModifiedDate.Compare(b.LastModifiedDate)
	})
	return subdirs
}

func (d *Directory) NameSortedSubDirs(comparator func(a, b string) int) []SubDirEntry {
	subdirs := slices.Collect(maps.Values(d.SubDirs))
	slices.SortFunc(subdirs, func(a, b SubDirEntry) int {
		return comparator(a.Name, b.Name)
	})
	return subdirs
}

// Returns the files contained within the directory as a slice, sorted
func (d *Directory) ModifiedTimeSortedFiles() []File {
	subdirs := slices.Collect(maps.Values(d.Files))
	slices.SortFunc(subdirs, func(a, b File) int {
		return a.LastModifiedDate.Compare(b.LastModifiedDate)
	})
	return subdirs
}

func (d *Directory) NameSortedFiles(comparator func(a, b string) int) []File {
	subdirs := slices.Collect(maps.Values(d.Files))
	slices.SortFunc(subdirs, func(a, b File) int {
		return comparator(a.Name, b.Name)
	})
	return subdirs
}

// The metadata representing a subdirectory
type SubDirEntry struct {
	// The name of the subdirectory
	Name string
	// The URL that this mirror subdirectory is contained at. This should not be accessed directly in most cases, with the Fetch function being preferable
	URL string
	// The date the directory was last modified, as reported by the mirror
	LastModifiedDate time.Time
}

func (s *SubDirEntry) Fetch(c Client) (*Directory, error) {
	dir, err := c.ReadDir(s.URL)
	if err != nil {
		return nil, err
	}
	dir.Name = s.Name
	dir.URL = s.URL
	return dir, nil
}

// File metadata
type File struct {
	// The name of the file
	Name string
	// The absolute URL of the file
	URL string
	// The time at which the file was last modified, as reported by the mirror
	LastModifiedDate time.Time
	// The size of the file, as reported by the mirror
	FileSize int64
}

type Client interface {
	ReadDir(url string) (*Directory, error)
}
