package os

import (
	"fmt"
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/mirror"
	"github.com/quickemu-project/quickget_configs/internal/web"
	"github.com/quickemu-project/quickget_configs/pkg/quickgetdata"
)

const launchpadReleasesUrl = "https://api.launchpad.net/devel/ubuntu/series"

var Edubuntu = OS{
	Name:        "edubuntu",
	PrettyName:  "Edubuntu",
	Homepage:    "https://www.edubuntu.org/",
	Description: "Stable, secure and privacy concious option for schools.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("edubuntu", x86_64_only[:], errs, csErrs)
	},
}

var Kubuntu = OS{
	Name:        "kubuntu",
	PrettyName:  "Kubuntu",
	Homepage:    "https://kubuntu.org/",
	Description: "Free, complete, and open-source alternative to Microsoft Windows and Mac OS X which contains everything you need to work, play, or share.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("kubuntu", x86_64_only[:], errs, csErrs)
	},
}

var Lubuntu = OS{
	Name:        "lubuntu",
	PrettyName:  "Lubuntu",
	Homepage:    "https://lubuntu.me/",
	Description: "Complete Operating System that ships the essential apps and services for daily use: office applications, PDF reader, image editor, music and video players, etc.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("lubuntu", x86_64_only[:], errs, csErrs)
	},
}

var Ubuntu = OS{
	Name:        "ubuntu",
	PrettyName:  "Ubuntu",
	Homepage:    "https://www.ubuntu.com/",
	Description: "Complete desktop Linux operating system, freely available with both community and professional support.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntu", x86_64_aarch64[:], errs, csErrs)
	},
}

var getUbuntuReleases = sync.OnceValues(fetchUbuntuReleases)

func fetchUbuntuReleases() ([]string, error) {
	var entries launchpadContents
	if err := web.CapturePageToJson(launchpadReleasesUrl, &entries); err != nil {
		return nil, err
	}

	releases := make([]string, 0)
	for _, e := range entries.Entries {
		if e.Status == "Supported" || e.Status == "Current Stable Release" {
			releases = append(releases, e.Version)
		}
	}

	return append(releases, "daily-live"), nil
}

var UbuntuBudgie = OS{
	Name:        "ubuntu-budgie",
	PrettyName:  "Ubuntu Budgie",
	Homepage:    "https://ubuntubudgie.org/",
	Description: "Community developed distribution, integrating the Budgie Desktop Environment with Ubuntu at its core.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntu-budgie", x86_64_only[:], errs, csErrs)
	},
}

var UbuntuCinnamon = OS{
	Name:        "ubuntu-cinnamon",
	PrettyName:  "Ubuntu Cinnamon",
	Homepage:    "https://ubuntucinnamon.org/",
	Description: "Community-driven, featuring Linux Mint's Cinnamon Desktop with Ubuntu at the core, packed fast and full of features, here is the most traditionally modern desktop you will ever love.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntucinnamon", x86_64_only[:], errs, csErrs)
	},
}

var UbuntuKylin = OS{
	Name:        "ubuntu-kylin",
	PrettyName:  "Ubuntu Kylin",
	Homepage:    "https://www.ubuntukylin.com/",
	Description: "Universal desktop operating system for personal computers, laptops, and embedded devices. It is dedicated to bringing a smarter user experience to users all over the world.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntukylin", x86_64_only[:], errs, csErrs)
	},
}

var UbuntuMATE = OS{
	Name:        "ubuntu-mate",
	PrettyName:  "Ubuntu MATE",
	Homepage:    "https://ubuntu-mate.org/",
	Description: "Stable, easy-to-use operating system with a configurable desktop environment. It is ideal for those who want the most out of their computers and prefer a traditional desktop metaphor.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntu-mate", x86_64_only[:], errs, csErrs)
	},
}

var UbuntuServer = OS{
	Name:        "ubuntu-server",
	PrettyName:  "Ubuntu Server",
	Homepage:    "https://www.ubuntu.com/server",
	Description: "Brings economic and technical scalability to your datacentre, public or private. Whether you want to deploy an OpenStack cloud, a Kubernetes cluster or a 50,000-node render farm, Ubuntu Server delivers the best value scale-out performance available.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntu-server", three_architectures[:], errs, csErrs)
	},
}

var UbuntuStudio = OS{
	Name:        "ubuntu-studio",
	PrettyName:  "Ubuntu Studio",
	Homepage:    "https://ubuntustudio.org/",
	Description: "Comes preinstalled with a selection of the most common free multimedia applications available, and is configured for best performance for various purposes: Audio, Graphics, Video, Photography and Publishing.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntustudio", x86_64_only[:], errs, csErrs)
	},
}

var UbuntuUnity = OS{
	Name:        "ubuntu-unity",
	PrettyName:  "Ubuntu Unity",
	Homepage:    "https://ubuntuunity.org/",
	Description: "Flavor of Ubuntu featuring the Unity7 desktop environment (the default desktop environment used by Ubuntu from 2010-2017).",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("ubuntu-unity", x86_64_only[:], errs, csErrs)
	},
}

var Xubuntu = OS{
	Name:        "xubuntu",
	PrettyName:  "Xubuntu",
	Homepage:    "https://xubuntu.org/",
	Description: "Elegant and easy to use operating system. Xubuntu comes with Xfce, which is a stable, light and configurable desktop environment.",
	ConfigFunction: func(errs, csErrs chan<- Failure) ([]Config, error) {
		return getUbuntuConfigs("xubuntu", x86_64_only[:], errs, csErrs)
	},
}

type launchpadContents struct {
	Entries []struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	} `json:"entries"`
}

func getUbuntuConfigs(variant string, architectures []Arch, errs, csErrs chan<- Failure) ([]Config, error) {
	releases, err := getUbuntuReleases()
	if err != nil {
		return nil, err
	}
	ch, wg := getChannels()

	for _, release := range releases {
		for _, arch := range architectures {
			wg.Go(func() {
				config, err, csErr := getUbuntuConfig(release, variant, arch)
				if err != nil {
					errs <- Failure{Release: release, Arch: arch, Error: err}
				}
				if csErr != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: csErr}
				}
				if config != nil {
					ch <- *config
				}
			})
		}
	}

	return waitForConfigs(ch, wg), nil
}

func getUbuntuConfig(release string, variant string, arch Arch) (config *Config, err error, csErr error) {
	c := mirror.HttpClient{}
	url := getUbuntuUrl(release, variant, arch)

	head, err := c.ReadDir(url)
	if err != nil {
		return nil, err, nil
	}

	checksums := make(map[string]string)
	if f, e := head.Files["SHA256SUMS"]; e {
		checksums, csErr = cs.Build(cs.Whitespace, f)
	}

	archText := getUbuntuArchText(arch)
	sku := getUbuntuSku(variant)
	var f *mirror.File
	for k, file := range head.Files {
		if strings.Contains(k, archText) && strings.Contains(k, sku) {
			f = &file
			break
		}
	}

	if f == nil {
		return
	}

	checksum := checksums["*"+f.Name]

	config = &Config{
		Release: release,
		Arch:    arch,
	}
	if !strings.Contains(release, "daily") && semverCompare(release, "16.04") < 0 {
		config.GuestOS = quickgetdata.LinuxOld
	}
	if arch == riscv64 {
		config.IMG = []Source{
			webSource(f.URL.String(), checksum, quickgetdata.Gz, f.Name),
		}
	} else {
		config.ISO = []Source{
			webSource(f.URL.String(), checksum, "", f.Name),
		}
	}

	return
}

func getUbuntuSku(variant string) string {
	switch variant {
	case "ubuntu-server":
		return "live-server"
	case "ubuntustudio":
		return "dvd"
	}
	return "desktop"
}

func getUbuntuArchText(arch Arch) string {
	switch arch {
	case x86_64:
		return "amd64.iso"
	case aarch64:
		return "arm64.iso"
	case riscv64:
		return "riscv64.img.gz"
	}
	return ""
}

func getUbuntuUrl(release, variant string, arch Arch) string {
	switch {
	case release == "daily-live":
		return fmt.Sprintf("https://cdimage.ubuntu.com/%s/daily-live/current/", variant)
	case arch == x86_64 && variant == "ubuntu" || variant == "ubuntu-server":
		return fmt.Sprintf("https://releases.ubuntu.com/%s/", release)
	case variant == "ubuntu-server":
		return fmt.Sprintf("https://cdimage.ubuntu.com/releases/%s/release/", release)
	}
	return fmt.Sprintf("https://cdimage.ubuntu.com/%s/releases/%s/release/", variant, release)
}
