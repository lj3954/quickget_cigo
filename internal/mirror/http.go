package mirror

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

// This regex should be changed to be more broad with dates when necessary.
// The date pattern should remain restrictive, since we will need to pass specific known layouts to time.Parse
var (
	preHttpMirrorReParts = []string{
		`<a href="([^"]+)">`,              // href url (match[1])
		`([^<]+)</a>`,                     // Name (match[2])
		`\s+`,                             // Padding
		`(\d{2}-\w{3}-\d{4} \d{2}:\d{2})`, // Date (match[3])
		`\s+`,                             // Padding
		`((?:[\d\.]+[BKMGT]?)|-)`,         // File size, or '-' for the absence of one (match[4])
	}
	preHttpMirrorRe = regexp.MustCompile(strings.Join(preHttpMirrorReParts, ""))

	timeLayouts = []string{
		"02-Jan-2006 15:04",
	}

	units = map[string]float64{
		"B": 1,
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
		"T": 1024 * 1024 * 1024 * 1024,
	}
)

func parseDate(value string) (t time.Time, err error) {
	for _, f := range timeLayouts {
		t, err = time.Parse(f, value)
		if err == nil {
			return
		}
	}
	return
}

func parseFileSize(value string) (int64, error) {
	if value == "-" {
		return -1, nil
	}

	nonDigitIndex := len(value)
	for nonDigitIndex > 0 {
		b := value[nonDigitIndex-1]
		if b >= '0' && b <= '9' {
			break
		}
		nonDigitIndex--
	}

	fmt.Println(nonDigitIndex, value)
	v, err := strconv.ParseFloat(value[:nonDigitIndex], 64)
	if err != nil {
		return 0, err
	}

	m, e := units[value[nonDigitIndex:]]
	if !e {
		m = 1
	}

	return int64(v * m), nil
}

type HttpMirrorClient struct{}

func (HttpMirrorClient) ReadDir(rawURL string) (*Directory, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	name := path.Base(u.Path)

	res, err := web.GetResponse(rawURL, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	text, err := doc.Find("pre").First().Html()
	if err != nil {
		return nil, err
	}

	matches := preHttpMirrorRe.FindAllStringSubmatch(text, -1)

	files := make(map[string]File)
	subdirs := make(map[string]SubDirEntry)
	for _, match := range matches {
		date, err := parseDate(match[3])
		if err != nil {
			return nil, err
		}
		name := match[2]
		url := rawURL + match[1]

		if name[len(name)-1] == '/' {
			name = name[:len(name)-1]
			subdirs[name] = SubDirEntry{
				Name:             name,
				URL:              url,
				LastModifiedDate: date,
			}
		} else {
			fileSize, err := parseFileSize(match[4])
			if err != nil {
				return nil, err
			}
			files[name] = File{
				Name:             name,
				URL:              url,
				LastModifiedDate: date,
				FileSize:         fileSize,
			}
		}
	}

	return &Directory{
		Name:    name,
		URL:     rawURL,
		Files:   files,
		SubDirs: subdirs,
	}, nil
}
