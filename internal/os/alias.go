package os

import (
	"github.com/quickemu-project/quickget_configs/internal/utils"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

type OSData utils.OSData
type Config utils.Config
type Arch quickgetdata.Arch

const x86_64 = quickgetdata.X86_64
const aarch64 = quickgetdata.Aarch64
const riscv64 = quickgetdata.Riscv64

var capturePage = utils.CapturePage
