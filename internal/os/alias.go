package os

import (
	"github.com/quickemu-project/quickget_configs/internal/utils"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

type OSData = utils.OSData
type Config = utils.Config
type Arch = qgdata.Arch
type Source = qgdata.Source

const x86_64 = qgdata.X86_64
const aarch64 = qgdata.Aarch64
const riscv64 = qgdata.Riscv64

var webSource = qgdata.NewWebSource
var urlChecksumSource = qgdata.URLChecksumSource
var urlSource = qgdata.URLSource

var capturePage = utils.CapturePage
var buildChecksum = utils.BuildChecksum
var singleWhitespaceChecksum = utils.SingleWhitespaceChecksum
var buildSingleWhitespaceChecksum = utils.BuildSingleWhitespaceChecksum
var getChannels = utils.GetChannels
var waitForConfigs = utils.WaitForConfigs

type Whitespace = utils.Whitespace
type Md5Regex = utils.Md5Regex
type Sha256Regex = utils.Sha256Regex
type CustomRegex = utils.CustomRegex
