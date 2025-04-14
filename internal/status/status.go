package status

import (
	"context"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/quickemu-project/quickget_configs/internal/data"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
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

func Create(len int) *Status {
	return &Status{
		StartTime: time.Now(),
		Data:      make([]osStatus, 0, len),
	}
}

func makeOsStatus(data qgdata.OSData) osStatus {
	return osStatus{
		Name:        data.Name,
		PrettyName:  data.PrettyName,
		Homepage:    data.Homepage,
		Description: data.Description,
	}
}

func (s *Status) FailedOS(data qgdata.OSData, err error) {
	s.Lock()
	defer s.Unlock()
	status := makeOsStatus(data)
	status.Err = err
	s.Data = append(s.Data, status)
}

func (s *Status) AddOS(data qgdata.OSData, failures, csFailures []data.Failure) {
	s.Lock()
	defer s.Unlock()
	status := makeOsStatus(data)
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
	for _, config := range data.Releases {
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

	page := statusTempl(s)
	if err := os.RemoveAll(statusPageDir); err != nil {
		return err
	} else if err := os.Mkdir(statusPageDir, 0755); err != nil {
		return err
	}
	file, err := os.Create(statusPageDir + "/index.html")
	if err != nil {
		return err
	}
	defer file.Close()
	if err := page.Render(context.Background(), file); err != nil {
		return err
	}

	return nil
}
