package os

type BlendOS struct{}

func (BlendOS) Data() OSData {
	return OSData{
		Name:        "blendos",
		PrettyName:  "BlendOS",
		Homepage:    "https://blendos.co/",
		Description: "A seamless blend of all Linux distributions. Allows you to have an immutable, atomic and declarative Arch Linux system, with application support from several Linux distributions & Android.",
	}
}

func (BlendOS) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return []Config{
		{
			ISO: []Source{
				urlSource("https://kc1.mirrors.199693.xyz/blend/isos/testing/blendOS.iso"),
			},
		},
	}, nil
}
