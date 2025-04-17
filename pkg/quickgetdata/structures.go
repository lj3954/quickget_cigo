package quickgetdata

type OSData struct {
	Name        string   `json:"name"`
	PrettyName  string   `json:"pretty_name"`
	Homepage    string   `json:"homepage"`
	Description string   `json:"description"`
	Releases    []Config `json:"releases"`
}

type Arch string

const (
	X86_64  Arch = "x86_64"
	Aarch64 Arch = "aarch64"
	Riscv64 Arch = "riscv64"
)

type GuestOS string

const (
	Linux         GuestOS = "linux"
	LinuxOld      GuestOS = "linux_old"
	Windows       GuestOS = "windows"
	WindowsServer GuestOS = "windows_server"
	MacOS         GuestOS = "macos"
	FreeBSD       GuestOS = "freebsd"
	GhostBSD      GuestOS = "ghostbsd"
	GenericBSD    GuestOS = "bsd"
	FreeDOS       GuestOS = "freedos"
	Haiku         GuestOS = "haiku"
	Solaris       GuestOS = "solaris"
	KolibriOS     GuestOS = "kolibrios"
	ReactOS       GuestOS = "reactos"
	Batocera      GuestOS = "batocera"
)

type DiskFormat string

const (
	Qcow2 DiskFormat = "qcow2"
	Raw   DiskFormat = "raw"
	Qed   DiskFormat = "qed"
	Qcow  DiskFormat = "qcow"
	Vdi   DiskFormat = "vdi"
	Vpc   DiskFormat = "vpc"
	Vhdx  DiskFormat = "vhdx"
)

type ArchiveFormat string

const (
	Tar      ArchiveFormat = "tar"
	TarBz2   ArchiveFormat = "tar.bz2"
	TarGz    ArchiveFormat = "tar.gz"
	TarXz    ArchiveFormat = "tar.xz"
	Xz       ArchiveFormat = "xz"
	Gz       ArchiveFormat = "gz"
	Bz2      ArchiveFormat = "bz2"
	Zip      ArchiveFormat = "zip"
	SevenZip ArchiveFormat = "7z"
)

type Config struct {
	Release    string   `json:"release"`
	Edition    string   `json:"edition,omitempty"`
	GuestOS    GuestOS  `json:"os,omitempty"`
	Arch       Arch     `json:"arch,omitempty"`
	ISO        []Source `json:"iso,omitempty"`
	IMG        []Source `json:"img,omitempty"`
	FixedISO   []Source `json:"fixed_iso,omitempty"`
	Floppy     []Source `json:"floppy,omitempty"`
	DiskImages []Disk   `json:"disk_images,omitempty"`
	TPM        bool     `json:"tpm,omitempty"`
	RAM        int64    `json:"ram,omitempty"`
	// This field tells the config generation to modify URL validation logic. This can be done because of ratelimits, datacenter IP blocking, or any other reason
	Validation Validation `json:"-"`
}

type Validation struct {
	Skip      bool
	Accept403 bool
}

type Disk struct {
	Source Source     `json:"source"`
	Size   int64      `json:"size,omitempty"`
	Format DiskFormat `json:"format,omitempty"`
}

type Source struct {
	Web      *WebSource    `json:"web,omitempty"`
	FileName string        `json:"file_name,omitempty"`
	Custom   bool          `json:"custom,omitempty"`
	Docker   *DockerSource `json:"docker,omitempty"`
}

type DockerSource struct {
	URL            string   `json:"url"`
	Privileged     bool     `json:"privileged,omitempty"`
	SharedDirs     []string `json:"shared_dirs,omitempty"`
	OutputFilename string   `json:"output_filename,omitempty"`
}

type WebSource struct {
	URL           string        `json:"url"`
	Checksum      string        `json:"checksum,omitempty"`
	ArchiveFormat ArchiveFormat `json:"archive_format,omitempty"`
	FileName      string        `json:"file_name,omitempty"`
}
