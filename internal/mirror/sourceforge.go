package mirror

import (
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/quickemu-project/quickget_configs/internal/web"
)

type SourceForgeClient struct{}

var (
	// There's HTML with all of the data we currently use, so there's no need at this time to parse this script
	//
	// sourceforgeTopLevelRe     = regexp.MustCompile(`net\.sf\.files\s+=\s+{(.*?)};`)
	// sourceforgeElementReParts = []string{
	// 	`"([^"]+)":{`,
	// 	`"name":"([^"]+)",`,
	// 	`"path":"([^"]*)",`,
	// 	`"download_url":"([^"]*)",`,
	// 	`"url":"([^"]*)",`,
	// 	`"full_path":"([^"]*)",`,
	// 	`"type":"(d|f)",`,
	// 	`"link":"([^"]*)",`,
	// 	`"downloads":(\d+),`,
	// 	`"sha1":"([^"]*)",`,
	// 	`"md5":"([^"]*)",`,
	// 	`"default":"([^"]*)",`,
	// 	`"download_label":"([^"]*)",`,
	// 	`"exclude_reports":(true|false),`,
	// 	`"downloadable":(true|false),`,
	// 	`"legacy_release_notes":(null|"[^"]*"),`,
	// 	`"staged":(true|false),`,
	// 	`"stage":(\d+),`,
	// 	`"staging_days":(\d+),`,
	// 	`"files_url":"([^"]+)",`,
	// 	`"explicitly_staged":(true|false),`,
	// 	`"authorized":(null|"[^"]*")`,
	// 	`}`,
	// }
	//sourceforgeElementRe = regexp.MustCompile(strings.Join(sourceforgeElementReParts, ""))

	sourceforgeTopLevelUrl, _ = url.Parse("https://sourceforge.net/")
	sourceforgeTimeFormat     = time.DateTime + " MST"
)

func (c SourceForgeClient) ReadDir(urlStr string) (*Directory, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return c.ReadDirFromUrl(u)
}

func (c SourceForgeClient) ReadDirFromUrl(u *url.URL) (*Directory, error) {
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

	doc.Find("table#files_list tbody tr").Each(func(i int, s *goquery.Selection) {
		classAttr, _ := s.Attr("class")
		classes := strings.Fields(classAttr)

		isFile := slices.Contains(classes, "file")
		isSubDir := slices.Contains(classes, "folder")
		if !isFile && !isSubDir {
			return
		}

		name := strings.TrimSpace(s.Find(".name").Text())
		rawUrl, _ := s.Find("th a").Attr("href")
		dateModifiedStr, _ := s.Find("td.opt abbr").Attr("title")
		dateModified, _ := time.Parse(sourceforgeTimeFormat, dateModifiedStr)

		if isSubDir {
			url := sourceforgeTopLevelUrl.JoinPath(rawUrl)
			subdirs[name] = SubDirEntry{
				client:           c,
				Name:             name,
				URL:              url,
				LastModifiedDate: dateModified,
			}
		} else {
			fileSizeStr := strings.TrimSpace(s.Find("td[headers='files_size_h']").Text())
			fileSize, _ := parseFileSize(fileSizeStr)
			url, err := url.Parse(rawUrl)
			if err != nil {
				return
			}
			files[name] = File{
				Name:             name,
				URL:              url,
				LastModifiedDate: dateModified,
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
