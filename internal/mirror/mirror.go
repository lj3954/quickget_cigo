package mirror

import (
	"iter"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"time"
)

type Directory struct {
	Name    string
	URL     *url.URL
	Files   map[string]File
	SubDirs map[string]SubDirEntry
}

// Returns the subdirectories contained within the directory as a slice, sorted by time
func (d *Directory) ModifiedTimeSortedSubdirs() []SubDirEntry {
	subdirs := slices.Collect(maps.Values(d.SubDirs))
	slices.SortFunc(subdirs, func(a, b SubDirEntry) int {
		return a.LastModifiedDate.Compare(b.LastModifiedDate)
	})
	return subdirs
}

// Returns the subdirectories contained within the directory as a slice, sorted by name with the provided comparator
func (d *Directory) NameSortedSubDirs(comparator func(a, b string) int) []SubDirEntry {
	subdirs := slices.Collect(maps.Values(d.SubDirs))
	slices.SortFunc(subdirs, func(a, b SubDirEntry) int {
		return comparator(a.Name, b.Name)
	})
	return subdirs
}

// Returns the subdirectories contained within the directory that match the given pattern
func (d *Directory) MatchingSubDirs(pattern *regexp.Regexp) iter.Seq[SubDirEntry] {
	return func(yield func(SubDirEntry) bool) {
		for k, d := range d.SubDirs {
			if pattern.MatchString(k) {
				if !yield(d) {
					return
				}
			}
		}
	}
}

// Returns the subdirectories contained within the directory that match the given pattern, and the produced match groups
func (d *Directory) SubDirMatches(pattern *regexp.Regexp) iter.Seq2[SubDirEntry, []string] {
	return func(yield func(SubDirEntry, []string) bool) {
		for k, d := range d.SubDirs {
			matches := pattern.FindStringSubmatch(k)
			if matches != nil {
				if !yield(d, matches) {
					return
				}
			}
		}
	}
}

// Returns the first subdirectory that a function returns true given, as well as whether the subdirectory exists.
func (d *Directory) FindSubDir(f func(s SubDirEntry) bool) (subdir SubDirEntry, ok bool) {
	for _, subdir := range d.SubDirs {
		if f(subdir) {
			return subdir, true
		}
	}
	return
}

// Returns the files contained within the directory as a slice, sorted by time
func (d *Directory) ModifiedTimeSortedFiles() []File {
	files := slices.Collect(maps.Values(d.Files))
	slices.SortFunc(files, func(a, b File) int {
		return a.LastModifiedDate.Compare(b.LastModifiedDate)
	})
	return files
}

// Returns the files contained within the directory as a slice, sorted by name with the provided comparator
func (d *Directory) NameSortedFiles(comparator func(a, b string) int) []File {
	files := slices.Collect(maps.Values(d.Files))
	slices.SortFunc(files, func(a, b File) int {
		return comparator(a.Name, b.Name)
	})
	return files
}

// Returns the files contained within the directory that match the given pattern
func (d *Directory) MatchingFiles(pattern *regexp.Regexp) iter.Seq[File] {
	return func(yield func(File) bool) {
		for k, f := range d.Files {
			if pattern.MatchString(k) {
				if !yield(f) {
					return
				}
			}
		}
	}
}

// Returns the files contained within the directory that match the given pattern, and the produced match groups
func (d *Directory) FileMatches(pattern *regexp.Regexp) iter.Seq2[File, []string] {
	return func(yield func(File, []string) bool) {
		for k, f := range d.Files {
			matches := pattern.FindStringSubmatch(k)
			if matches != nil {
				if !yield(f, matches) {
					return
				}
			}
		}
	}
}

// Returns the first file that a function returns true given, as well as whether the file exists.
func (d *Directory) FindFile(f func(f File) bool) (file File, ok bool) {
	for _, file := range d.Files {
		if f(file) {
			return file, true
		}
	}
	return
}

// The metadata representing a subdirectory
type SubDirEntry struct {
	// The internal client used to fetch this subdirectory (the client type from the initial directory read)
	client Client
	// The name of the subdirectory
	Name string
	// The URL that this mirror subdirectory is contained at. This should not be accessed directly in most cases, with the Fetch function being preferable
	URL *url.URL
	// The date the directory was last modified, as reported by the mirror
	LastModifiedDate time.Time
}

func (s *SubDirEntry) Fetch() (*Directory, error) {
	dir, err := s.client.ReadDirFromUrl(s.URL)
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
	URL *url.URL
	// The time at which the file was last modified, as reported by the mirror
	LastModifiedDate time.Time
	// The size of the file, as reported by the mirror
	FileSize int64
}

type Client interface {
	// This method should be implemented to parse the url string and then call the ReadDirFromUrl method on self
	ReadDir(urlStr string) (*Directory, error)
	// Read a directory given a URL. Returns an error if the client fails to make an HTTP request or parse the resulting data
	ReadDirFromUrl(u *url.URL) (*Directory, error)
}
