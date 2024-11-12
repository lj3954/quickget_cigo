package os

import (
	"github.com/quickemu-project/quickget_configs/internal/utils"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

type OSData = utils.OSData
type Config = utils.Config
type GithubAPI = utils.GithubAPI
type GithubAsset = utils.GithubAsset
type Arch = qgdata.Arch
type ArchiveFormat = qgdata.ArchiveFormat
type Source = qgdata.Source
type Disk = qgdata.Disk
type Failure = utils.Failure

const x86_64 = qgdata.X86_64
const aarch64 = qgdata.Aarch64
const riscv64 = qgdata.Riscv64

const Gz = qgdata.Gz

var webSource = qgdata.NewWebSource
var urlChecksumSource = qgdata.URLChecksumSource
var urlSource = qgdata.URLSource

var capturePage = utils.CapturePage
var buildChecksum = utils.BuildChecksum
var singleWhitespaceChecksum = utils.SingleWhitespaceChecksum
var buildSingleWhitespaceChecksum = utils.BuildSingleWhitespaceChecksum
var getChannels = utils.GetChannels
var waitForConfigs = utils.WaitForConfigs
var getBasicReleases = utils.GetBasicReleases

type Whitespace = utils.Whitespace
type CustomRegex = utils.CustomRegex

var Md5Regex = utils.Md5Regex
var Sha256Regex = utils.Sha256Regex
