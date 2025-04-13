package os

import (
	"github.com/quickemu-project/quickget_configs/internal/data"
	"github.com/quickemu-project/quickget_configs/internal/utils"
	qgdata "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

type (
	OSData        = utils.OSData
	OS            = utils.OS
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

var List = []OS{
	alma,
	alpine,
	antiX,
	archCraft,
	archLinux,
	arcoLinux,
	artixLinux,
	athenaOS,
	azureLinux,
	batocera,
	bazzite,
	bigLinux,
	blendOS,
	bodhi,
	bunsenLabs,
	cachyOS,
	centOSStream,
	chimeraLinux,
	cbpp,
	debian,
	deepin,
	devuan,
	dragonFlyBSD,
	easyOS,
	edubuntu,
	elementary,
	endeavourOS,
	endlessOS,
	fedora,
	freeBSD,
	freeDOS,
	garuda,
	gentoo,
	ghostBSD,
	gnomeOS,
	guix,
	haiku,
	kali,
	kdeNeon,
	kolibriOS,
	kubuntu,
	linuxLite,
	linuxMint,
	lmde,
	lubuntu,
	mageia,
	manjaro,
	mxLinux,
	netboot,
	netBSD,
	nitrux,
	nixOS,
	nwgShell,
	openBSD,
	openIndiana,
	openSUSE,
	oracleLinux,
	parrotSec,
	peppermint,
	popOS,
	porteus,
	primtux,
	proxmoxVE,
	pureOS,
	reactOS,
	rebornOS,
	rockyLinux,
	siduction,
	slackware,
	slax,
	slint,
	slitaz,
	solus,
	ubuntu,
	ubuntuBudgie,
	ubuntuCinnamon,
	ubuntuKylin,
	ubuntuMATE,
	ubuntuServer,
	ubuntuStudio,
	ubuntuUnity,
	windowsServer,
	windows,
	xubuntu,
}
