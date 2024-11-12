package cs

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/quickemu-project/quickget_configs/internal/utils"
)

func SingleWhitespace(url string) (string, error) {
	data, err := utils.CapturePage(url)
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

func Build(cs ChecksumSeparation, url string) (map[string]string, error) {
	data, err := utils.CapturePage(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to build checksums: %w", err)
	}
	return cs.BuildWithData(data), nil
}

type Whitespace struct{}
type CustomRegex struct {
	Regex      *regexp.Regexp
	KeyIndex   int
	ValueIndex int
}

var Md5Regex = CustomRegex{
	Regex:      regexp.MustCompile(`MD5 \(([^)]+)\) = ([0-9a-f]+)`),
	KeyIndex:   1,
	ValueIndex: 2,
}
var Sha256Regex = CustomRegex{
	Regex:      regexp.MustCompile(`SHA256 \(([^)]+)\) = ([0-9a-f]+)`),
	KeyIndex:   1,
	ValueIndex: 2,
}

func (Whitespace) BuildWithData(data string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(data, "\n") {
		slice := strings.SplitN(line, " ", 2)
		if len(slice) == 2 {
			hash := strings.TrimSpace(slice[0])
			file := strings.TrimSpace(slice[1])
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
