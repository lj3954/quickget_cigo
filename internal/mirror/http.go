package mirror

import (
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

type mirrorClass int

const (
	mirrorClassNone = mirrorClass(iota)
	mirrorClassLink
	mirrorClassFileSize
	mirrorClassDateModified
)

var (
	// This regex should be changed to be more broad with dates when necessary.
	// The date pattern should remain restrictive, since we will need to pass specific known layouts to time.Parse
	preHttpMirrorReParts = []string{
		`<a href="([^"]+)">`,                 // href url (match[1])
		`([^<]+)</a>`,                        // Name (match[2])
		`\s+`,                                // Padding
		`(`,                                  // Date group (match[3]) opening
		`(?:\d{2}-\w{3}-\d{4} \d{2}:\d{2})`,  // First date format
		`|(?:\d{4}-\d{2}-\d{2} \d{2}:\d{2})`, // Second date format
		`)`,                                  // Closing date group
		`\s+`,                                // Padding
		`((?:[\d\.]+[BKMGT]?)|-)`,            // File size, or '-' for the absence of one (match[4])
	}
	preHttpMirrorRe = regexp.MustCompile(strings.Join(preHttpMirrorReParts, ""))

	timeLayouts = []string{
		"02-Jan-2006 15:04",
		"2006-01-02 15:04",
		"2006-Jan-02 15:04",
	}

	units = map[string]float64{
		"B":      1,
		" Bytes": 1,
		"K":      1024,
		" kB":    1024,
		"M":      1024 * 1024,
		" MB":    1024 * 1024,
		"G":      1024 * 1024 * 1024,
		" GB":    1024 * 1024 * 1024,
		"T":      1024 * 1024 * 1024 * 1024,
		" TB":    1024 * 1024 * 1024 * 1024,
	}

	mirrorClasses = map[string]mirrorClass{
		"link": mirrorClassLink,
		"n":    mirrorClassLink,

		"size": mirrorClassFileSize,
		"s":    mirrorClassFileSize,

		"date": mirrorClassDateModified,
		"m":    mirrorClassDateModified,
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

	v, err := strconv.ParseFloat(value[:nonDigitIndex], 64)
	if err != nil {
		return 0, err
	}

	m, ok := units[value[nonDigitIndex:]]
	if !ok {
		m = 1
	}

	return int64(v * m), nil
}

type LegacyHttpClient struct{}

func (c LegacyHttpClient) ReadDir(urlStr string) (*Directory, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return c.ReadDirFromUrl(u)
}

func (c LegacyHttpClient) ReadDirFromUrl(u *url.URL) (*Directory, error) {
	name := path.Base(u.Path)

	res, err := web.GetResponse(u, nil)
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
		url := u.JoinPath(match[1])

		// Files ending with this pattern have a name cut off for length (...)
		// Use the final path segment as the name if we spot this pattern
		if strings.HasSuffix(name, "..&gt;") {
			name = path.Base(url.Path)
		}

		if strings.HasSuffix(match[1], "/") {
			name = strings.TrimSuffix(name, "/")
			subdirs[name] = SubDirEntry{
				client:           c,
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
		URL:     u,
		Files:   files,
		SubDirs: subdirs,
	}, nil
}

type HttpClient struct{}

func (c HttpClient) ReadDir(urlStr string) (*Directory, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return c.ReadDirFromUrl(u)
}

func (c HttpClient) ReadDirFromUrl(u *url.URL) (*Directory, error) {
	name := path.Base(u.Path)

	res, err := web.GetResponse(u, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	files := make(map[string]File)
	subdirs := make(map[string]SubDirEntry)

	doc.Find("table tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")

		if cells.Length() < 1 {
			return
		}

		var name, link string
		fileSize := int64(-1)
		var modifiedDate time.Time

		cells.Each(func(i int, s *goquery.Selection) {
			classAttr, _ := s.Attr("class")
			classes := strings.Fields(classAttr)

			var class mirrorClass
			for _, c := range classes {
				if v, ok := mirrorClasses[c]; ok {
					class = v
					break
				}
			}

			if class == mirrorClassNone {
				if a := s.Find("a"); a.Length() == 1 {
					_, ok := a.Attr("href")
					t := strings.TrimSpace(a.Text())
					if ok && len(t) > 0 {
						class = mirrorClassLink
					}
				} else if _, err := parseDate(s.Text()); err == nil {
					class = mirrorClassDateModified
				} else if _, err := parseFileSize(s.Text()); err == nil {
					class = mirrorClassFileSize
				}
			}

			switch class {
			case mirrorClassNone:
				return
			case mirrorClassLink:
				a := s.Find("a")
				href, ok := a.Attr("href")
				if !ok {
					return
				}
				link = href
				name = strings.TrimSpace(a.Text())
			case mirrorClassFileSize:
				if v, ok := s.Attr("data-value"); ok {
					size, err := strconv.ParseInt(v, 10, 64)
					if err == nil {
						fileSize = size
						return
					}
				}
				fileSize, _ = parseFileSize(s.Text())
			case mirrorClassDateModified:
				modifiedDate, _ = parseDate(s.Text())
			}
		})

		if len(name) == 0 || len(link) == 0 || link == "../" {
			return
		}

		url := u.JoinPath(link)

		if strings.HasSuffix(link, "/") {
			name = strings.TrimSuffix(name, "/")
			subdirs[name] = SubDirEntry{
				client:           c,
				Name:             name,
				URL:              url,
				LastModifiedDate: modifiedDate,
			}
		} else {
			files[name] = File{
				Name:             name,
				URL:              url,
				LastModifiedDate: modifiedDate,
				FileSize:         fileSize,
			}
		}
	})

	return &Directory{
		Name:    name,
		URL:     u,
		Files:   files,
		SubDirs: subdirs,
	}, nil
}
