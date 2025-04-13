package os

import "github.com/quickemu-project/quickget_configs/internal/cs"

var peppermint = OS{
	Name:           "peppermint",
	PrettyName:     "PeppermintOS",
	Homepage:       "https://peppermintos.com/",
	Description:    `Provides a user with the opportunity to build the system that best fits their needs. While at the same time providing a functioning OS with minimum hassle out of the box.`,
	ConfigFunction: createPeppermintConfigs,
}

// Peppermint's sourceforge is very convoluted and can't easily be parsed.
// Non-matching checksum & iso naming, etc
// Therefore, we'll just hardcode values (yes, this was done manually)

func createPeppermintConfigs(errs, csErrs chan<- Failure) ([]Config, error) {
	releases := []peppermintRelease{
		{
			release:     "debian",
			edition:     "xfce-loaded",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-Debian-64_loaded.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-DebianLoaded-64-sha512.checksum/download",
		},
		{
			release:     "debian",
			edition:     "xfce",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-Debian-64.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-Debian-64-sha512.checksum/download",
		},
		{
			release:     "debian",
			edition:     "xfce",
			arch:        aarch64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-Debian_ARM_64.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-Debian_ARM_64-sha512.checksum/download",
		},
		{
			release:     "devuan",
			edition:     "xfce-loaded",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-devuan_64_loaded.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-DevuanLoaded-64-sha512.checksum/download",
		},
		{
			release:     "devuan",
			edition:     "xfce",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-devuan_64_xfce.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-devuan_64_xfce-sha512.checksum/download",
		},
		{
			release:     "devuan",
			edition:     "xfce",
			arch:        aarch64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-devuan_arm_xfce.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/XFCE/PeppermintOS-devuan_arma_xfce-sha512.checksum/download",
		},
		{
			release:     "debian",
			edition:     "mini",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/Mini/PepMini-deb-amd64.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/Mini/PepMini-deb-amd64-sha512.checksum/download",
		},
		{
			release:     "devuan",
			edition:     "mini",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/Mini/PepMini-dev-amd64.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/Mini/PepMini-dev-amd64-sha512.checksum/download",
		},
		{
			release:     "debian",
			edition:     "gnome",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/Gnome_FlashBack/PeppermintOS-Debian_64_gfb.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/Gnome_FlashBack/PeppermintOS-Debian_64_gfb-sha512.checksum/download",
		},
		{
			release:     "devuan",
			edition:     "gnome",
			arch:        x86_64,
			url:         "https://sourceforge.net/projects/peppermintos/files/isos/Gnome_FlashBack/PeppermintOS-devuan_64_gfb.iso/download",
			checksumUrl: "https://sourceforge.net/projects/peppermintos/files/isos/Gnome_FlashBack/PeppermintOS-devuan_64_gnome_flashback-sha512.checksum/download",
		},
	}

	ch, wg := getChannelsWith(len(releases))
	for _, release := range releases {
		go func() {
			defer wg.Done()
			cs, err := cs.SingleWhitespace(release.checksumUrl)
			if err != nil {
				csErrs <- Failure{Release: release.release, Edition: release.edition, Error: err}
			}
			ch <- Config{
				Release: release.release,
				Edition: release.edition,
				Arch:    release.arch,
				ISO: []Source{
					urlChecksumSource(release.url, cs),
				},
			}
		}()
	}

	return waitForConfigs(ch, wg), nil
}

type peppermintRelease struct {
	release     string
	edition     string
	arch        Arch
	url         string
	checksumUrl string
}
