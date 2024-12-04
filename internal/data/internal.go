package data

import "github.com/quickemu-project/quickget_configs/pkg/quickgetdata"

type OSData struct {
	Name        string                `json:"name"`
	PrettyName  string                `json:"pretty_name"`
	Homepage    string                `json:"homepage"`
	Description string                `json:"description"`
	Releases    []quickgetdata.Config `json:"releases"`
}
