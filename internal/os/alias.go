package os

import (
	"github.com/quickemu-project/quickget_configs/internal/data"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

type (
	OSData        = utils.OSData
	Config        = utils.Config
	GithubAPI     = data.GithubAPI
	GithubAsset   = data.GithubAsset
	Arch          = qgdata.Arch
	ArchiveFormat = qgdata.ArchiveFormat
	Source        = qgdata.Source
	Disk          = qgdata.Disk
	Failure       = data.Failure
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
	getChannels           = utils.GetChannels
	getChannelsWith       = utils.GetChannelsWith
	waitForConfigs        = utils.WaitForConfigs
	getBasicReleases      = utils.GetBasicReleases
	getReverseReleases    = utils.GetReverseReleases
	getSortedReleases     = utils.GetSortedReleases
	getSortedReleasesFunc = utils.GetSortedReleasesFunc
	integerCompare        = utils.IntegerCompare
	semverCompare         = utils.SemverCompare
)

var (
	x86_64_only         = [...]Arch{x86_64}
	x86_64_aarch64      = [...]Arch{x86_64, aarch64}
	three_architectures = [...]Arch{x86_64, aarch64, riscv64}
)
