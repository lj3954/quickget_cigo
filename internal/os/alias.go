package os

import (
	"github.com/quickemu-project/quickget_configs/internal/utils"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

type (
	OSData        = utils.OSData
	Config        = utils.Config
	GithubAPI     = utils.GithubAPI
	GithubAsset   = utils.GithubAsset
	Arch          = qgdata.Arch
	ArchiveFormat = qgdata.ArchiveFormat
	Source        = qgdata.Source
	Disk          = qgdata.Disk
	Failure       = utils.Failure
)

const (
	x86_64  = qgdata.X86_64
	aarch64 = qgdata.Aarch64
	riscv64 = qgdata.Riscv64
)

var (
	webSource         = qgdata.NewWebSource
	urlChecksumSource = qgdata.URLChecksumSource
	urlSource         = qgdata.URLSource
)

var (
	capturePage       = utils.CapturePage
	capturePageToJson = utils.CapturePageToJson[any]
	getChannels       = utils.GetChannels
	waitForConfigs    = utils.WaitForConfigs
	getBasicReleases  = utils.GetBasicReleases
)

var (
	x86_64_only         = [...]Arch{x86_64}
	x86_64_aarch64      = [...]Arch{x86_64, aarch64}
	three_architectures = [...]Arch{x86_64, aarch64, riscv64}
)
