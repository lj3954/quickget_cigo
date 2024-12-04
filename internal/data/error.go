package data

import "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"

type Failure struct {
	Release string
	Edition string
	Arch    quickgetdata.Arch
	Error   error
}
