package utils

import (
	"embed"
	"html/template"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

type Status struct {
	sync.Mutex
	StartTime time.Time
	EndTime   time.Time
	Data      []osStatus
}

type osStatus struct {
	Name        string
	PrettyName  string
	Homepage    string
	Description string
	Releases    []ReleaseStatus
	Err         error
}

type ReleaseStatus struct {
	Release    string
	Edition    string
	Arch       qgdata.Arch
	Sources    []sourceData
	DiskImages []qgdata.Disk
	CsErrs     []error
	Err        error
}

type sourceData struct {
	SourceType string
	Source     qgdata.Source
}

func createStatus(len int) *Status {
	return &Status{
		StartTime: time.Now(),
		Data:      make([]osStatus, 0, len),
	}
}

func (data OSData) toStatus() osStatus {
	return osStatus{
		Name:        data.Name,
		PrettyName:  data.PrettyName,
		Homepage:    data.Homepage,
		Description: data.Description,
	}
}

func (s *Status) failedOS(data OSData, err error) {
	s.Lock()
	defer s.Unlock()
	status := data.toStatus()
	status.Err = err
	s.Data = append(s.Data, status)
}

func (s *Status) addOS(data OSData, configs []Config, failures, csFailures []Failure) {
	s.Lock()
	defer s.Unlock()
	status := data.toStatus()
	for _, failure := range failures {
		status.Releases = append(status.Releases, ReleaseStatus{
			Release: failure.Release,
			Edition: failure.Edition,
			Arch:    failure.Arch,
			Err:     failure.Error,
		})
	}
	for _, failure := range csFailures {
		log.Printf("(Unimplemented) Checksum Failure: %s", failure)
	}
	for _, config := range configs {
		sourceLen := len(config.ISO) + len(config.IMG) + len(config.FixedISO) + len(config.Floppy)
		sources := make([]sourceData, 0, sourceLen)
		addSources(&sources, "ISO", config.ISO)
		addSources(&sources, "IMG", config.IMG)
		addSources(&sources, "Fixed ISO (CD-ROM)", config.FixedISO)
		addSources(&sources, "Floppy", config.Floppy)

		status.Releases = append(status.Releases, ReleaseStatus{
			Release:    config.Release,
			Edition:    config.Edition,
			Arch:       config.Arch,
			Sources:    sources,
			DiskImages: config.DiskImages,
		})
	}
	s.Data = append(s.Data, status)
}

func addSources(data *[]sourceData, sourceType string, sources []qgdata.Source) {
	for _, source := range sources {
		*data = append(*data, sourceData{
			SourceType: sourceType,
			Source:     source,
		})
	}
}

const statusPageDir = "statuspage"

func (s *Status) Finalize() error {
	s.Lock()
	s.EndTime = time.Now()

	slices.SortFunc(s.Data, func(a, b osStatus) int {
		return strings.Compare(a.Name, b.Name)
	})
	html, err := statusTemplate.ReadFile("statusTemplate/template.html")
	if err != nil {
		return err
	}
	funcs := template.FuncMap{
		"div": func(a, b float64) float64 {
			return a / b
		},
	}
	htmlTemplate, err := template.New("status").Funcs(funcs).Parse(string(html))
	if err != nil {
		return err
	}
	if err := os.Mkdir(statusPageDir, 0755); err != nil {
		return err
	}
	file, err := os.Create(statusPageDir + "/index.html")
	if err != nil {
		return err
	}
	defer file.Close()
	if err := htmlTemplate.Execute(file, s); err != nil {
		return err
	}

	return nil
}

//go:embed statusTemplate/*
var statusTemplate embed.FS
