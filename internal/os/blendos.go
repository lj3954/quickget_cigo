package os

var BlendOS = OS{
	Name:           "blendos",
	PrettyName:     "BlendOS",
	Homepage:       "https://blendos.co/",
	Description:    "A seamless blend of all Linux distributions. Allows you to have an immutable, atomic and declarative Arch Linux system, with application support from several Linux distributions & Android.",
	ConfigFunction: createBlendOSConfigs,
}

func createBlendOSConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	return []Config{
		{
			ISO: []Source{
				urlSource("https://git.blendos.co/api/v4/projects/32/jobs/artifacts/main/raw/blendOS.iso?job=build-job"),
			},
		},
	}, nil
}
