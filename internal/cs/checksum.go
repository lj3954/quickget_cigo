package cs

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

func SingleWhitespace[T string | *url.URL | mirror.File](input T) (string, error) {
	var data string
	var err error

	switch v := any(input).(type) {
	case string:
		data, err = web.CapturePage(v)
	case *url.URL:
		data, err = web.CapturePage(v)
	case mirror.File:
		data, err = web.CapturePage(v.URL)
	}
	if err != nil {
		return "", fmt.Errorf("Failed to find single checksum: %w", err)
	}
	return BuildSingleWhitespace(data)
}

func BuildSingleWhitespace(data string) (string, error) {
	index := strings.Index(data, " ")
	if index == -1 {
		return "", errors.New("No whitespace was present in the checksum data")
	}
	return data[:index], nil
}

type ChecksumSeparation interface {
	BuildWithData(string) map[string]string
}

// Builds a checksum map from the contents of a URL and a pattern. Errors when the URL cannot be resolved.
// Return map is guaranteed to always be valid, even in the case of an error
func Build[T string | *url.URL | mirror.File](cs ChecksumSeparation, input T) (map[string]string, error) {
	var data string
	var err error

	switch v := any(input).(type) {
	case string:
		data, err = web.CapturePage(v)
	case *url.URL:
		data, err = web.CapturePage(v)
	case mirror.File:
		data, err = web.CapturePage(v.URL)
	}
	if err != nil {
		return make(map[string]string), fmt.Errorf("Failed to build checksums: %w", err)
	}
	return cs.BuildWithData(data), nil
}

type innerWhitespace struct{}

var Whitespace = innerWhitespace{}

type CustomRegex struct {
	Regex      *regexp.Regexp
	KeyIndex   int
	ValueIndex int
}

var Md5Regex = CustomRegex{
	Regex:      regexp.MustCompile(`MD5 \(([^)]+)\) = ([0-9a-f]{32})`),
	KeyIndex:   1,
	ValueIndex: 2,
}
var Sha256Regex = CustomRegex{
	Regex:      regexp.MustCompile(`SHA256 \(([^)]+)\) = ([0-9a-f]{64})`),
	KeyIndex:   1,
	ValueIndex: 2,
}
var Sha512Regex = CustomRegex{
	Regex:      regexp.MustCompile(`SHA512 \(([^)]+)\) = ([0-9a-f]{128})`),
	KeyIndex:   1,
	ValueIndex: 2,
}

func (innerWhitespace) BuildWithData(data string) map[string]string {
	m := make(map[string]string)
	for line := range strings.Lines(data) {
		slice := strings.SplitN(line, " ", 2)
		if len(slice) == 2 {
			hash := strings.TrimSpace(slice[0])
			file := path.Clean(strings.TrimSpace(slice[1]))
			m[file] = hash
		}
	}
	return m
}

func (re CustomRegex) BuildWithData(data string) map[string]string {
	m := make(map[string]string)
	for _, match := range re.Regex.FindAllStringSubmatch(data, -1) {
		file := match[re.KeyIndex]
		hash := match[re.ValueIndex]
		m[file] = hash
	}
	return m
}
