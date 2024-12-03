package os

import (
	"fmt"
	"strings"
	"sync"

	"github.com/quickemu-project/quickget_configs/internal/cs"
	"github.com/quickemu-project/quickget_configs/internal/web"
	quickgetdata "github.com/quickemu-project/quickget_configs/pkg/quickget_data"
)

const launchpadReleasesUrl = "https://api.launchpad.net/devel/ubuntu/series"

type (
	Edubuntu       struct{}
	Kubuntu        struct{}
	Lubuntu        struct{}
	Ubuntu         struct{}
	UbuntuBudgie   struct{}
	UbuntuCinnamon struct{}
	UbuntuKylin    struct{}
	UbuntuMATE     struct{}
	UbuntuServer   struct{}
	UbuntuStudio   struct{}
	UbuntuUnity    struct{}
	Xubuntu        struct{}
)

func (Edubuntu) Data() OSData {
	return OSData{
		Name:        "edubuntu",
		PrettyName:  "Edubuntu",
		Homepage:    "https://www.edubuntu.org/",
		Description: "Stable, secure and privacy concious option for schools.",
	}
}

func (Edubuntu) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("edubuntu", x86_64_only[:], errs, csErrs)
}

func (Kubuntu) Data() OSData {
	return OSData{
		Name:        "kubuntu",
		PrettyName:  "Kubuntu",
		Homepage:    "https://kubuntu.org/",
		Description: "Free, complete, and open-source alternative to Microsoft Windows and Mac OS X which contains everything you need to work, play, or share.",
	}
}

func (Kubuntu) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("kubuntu", x86_64_only[:], errs, csErrs)
}

func (Lubuntu) Data() OSData {
	return OSData{
		Name:        "lubuntu",
		PrettyName:  "Lubuntu",
		Homepage:    "https://lubuntu.me/",
		Description: "Complete Operating System that ships the essential apps and services for daily use: office applications, PDF reader, image editor, music and video players, etc.",
	}
}

func (Lubuntu) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("lubuntu", x86_64_only[:], errs, csErrs)
}

func (Ubuntu) Data() OSData {
	return OSData{
		Name:        "ubuntu",
		PrettyName:  "Ubuntu",
		Homepage:    "https://www.ubuntu.com/",
		Description: "Complete desktop Linux operating system, freely available with both community and professional support.",
	}
}

func (Ubuntu) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntu", x86_64_aarch64[:], errs, csErrs)
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

func (UbuntuBudgie) Data() OSData {
	return OSData{
		Name:        "ubuntu-budgie",
		PrettyName:  "Ubuntu Budgie",
		Homepage:    "https://ubuntubudgie.org/",
		Description: "Community developed distribution, integrating the Budgie Desktop Environment with Ubuntu at its core.",
	}
}

func (UbuntuBudgie) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntu-budgie", x86_64_only[:], errs, csErrs)
}

func (UbuntuCinnamon) Data() OSData {
	return OSData{
		Name:        "ubuntu-cinnamon",
		PrettyName:  "Ubuntu Cinnamon",
		Homepage:    "https://ubuntucinnamon.org/",
		Description: "Community-driven, featuring Linux Mint’s Cinnamon Desktop with Ubuntu at the core, packed fast and full of features, here is the most traditionally modern desktop you will ever love.",
	}
}

func (UbuntuCinnamon) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntucinnamon", x86_64_only[:], errs, csErrs)
}

func (UbuntuKylin) Data() OSData {
	return OSData{
		Name:        "ubuntu-kylin",
		PrettyName:  "Ubuntu Kylin",
		Homepage:    "https://www.ubuntukylin.com/",
		Description: "Universal desktop operating system for personal computers, laptops, and embedded devices. It is dedicated to bringing a smarter user experience to users all over the world.",
	}
}

func (UbuntuKylin) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntukylin", x86_64_only[:], errs, csErrs)
}

func (UbuntuMATE) Data() OSData {
	return OSData{
		Name:        "ubuntu-mate",
		PrettyName:  "Ubuntu MATE",
		Homepage:    "https://ubuntu-mate.org/",
		Description: "Stable, easy-to-use operating system with a configurable desktop environment. It is ideal for those who want the most out of their computers and prefer a traditional desktop metaphor.",
	}
}

func (UbuntuMATE) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntu-mate", x86_64_only[:], errs, csErrs)
}

func (UbuntuServer) Data() OSData {
	return OSData{
		Name:        "ubuntu-server",
		PrettyName:  "Ubuntu Server",
		Homepage:    "https://www.ubuntu.com/server",
		Description: "Brings economic and technical scalability to your datacentre, public or private. Whether you want to deploy an OpenStack cloud, a Kubernetes cluster or a 50,000-node render farm, Ubuntu Server delivers the best value scale-out performance available.",
	}
}

func (UbuntuServer) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntu-server", three_architectures[:], errs, csErrs)
}

func (UbuntuStudio) Data() OSData {
	return OSData{
		Name:        "ubuntu-studio",
		PrettyName:  "Ubuntu Studio",
		Homepage:    "https://ubuntustudio.org/",
		Description: "Comes preinstalled with a selection of the most common free multimedia applications available, and is configured for best performance for various purposes: Audio, Graphics, Video, Photography and Publishing.",
	}
}

func (UbuntuStudio) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntustudio", x86_64_only[:], errs, csErrs)
}

func (UbuntuUnity) Data() OSData {
	return OSData{
		Name:        "ubuntu-unity",
		PrettyName:  "Ubuntu Unity",
		Homepage:    "https://ubuntuunity.org/",
		Description: "Flavor of Ubuntu featuring the Unity7 desktop environment (the default desktop environment used by Ubuntu from 2010-2017).",
	}
}

func (UbuntuUnity) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("ubuntu-unity", x86_64_only[:], errs, csErrs)
}

func (Xubuntu) Data() OSData {
	return OSData{
		Name:        "xubuntu",
		PrettyName:  "Xubuntu",
		Homepage:    "https://xubuntu.org/",
		Description: "Elegant and easy to use operating system. Xubuntu comes with Xfce, which is a stable, light and configurable desktop environment.",
	}
}

func (Xubuntu) CreateConfigs(errs, csErrs chan Failure) ([]Config, error) {
	return getUbuntuConfigs("xubuntu", x86_64_only[:], errs, csErrs)
}

type launchpadContents struct {
	Entries []struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	} `json:"entries"`
}

func getUbuntuConfigs(variant string, architectures []Arch, errs, csErrs chan Failure) ([]Config, error) {
	releases, err := getUbuntuReleases()
	if err != nil {
		return nil, err
	}
	ch, wg := getChannelsWith(len(releases) * len(architectures))
	for _, release := range releases {
		for _, arch := range architectures {
			url := getUbuntuUrl(release, variant, arch)
			go func() {
				defer wg.Done()
				page, err := web.CapturePage(url + "SHA256SUMS")
				if err != nil {
					errs <- Failure{Release: release, Arch: arch, Error: err}
					return
				}
				line := getUbuntuLine(page, variant, arch)
				if line == "" {
					return
				}

				checksum, err := cs.BuildSingleWhitespace(line)
				if err != nil {
					csErrs <- Failure{Release: release, Arch: arch, Error: err}
				}
				iso := url + line[strings.Index(line, "*")+1:]

				config := Config{
					Release: release,
					Arch:    arch,
				}
				if !strings.Contains(release, "daily") && semverCompare(release, "16.04") < 0 {
					config.GuestOS = quickgetdata.LinuxOld
				}
				if arch == riscv64 {
					config.IMG = []Source{
						webSource(iso, checksum, quickgetdata.Gz, ""),
					}
				} else {
					config.ISO = []Source{
						urlChecksumSource(iso, checksum),
					}
				}

				ch <- config
			}()
		}
	}

	return waitForConfigs(ch, wg), nil
}

func getUbuntuLine(page, variant string, arch Arch) string {
	archText := getUbuntuArchText(arch)
	sku := getUbuntuSku(variant)
	for _, l := range strings.Split(page, "\n") {
		if strings.Contains(l, archText) && strings.Contains(l, sku) {
			return l
		}
	}
	return ""
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
