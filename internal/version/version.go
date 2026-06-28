package version

const (
	Name = "kiri"
	Version = "v0.1.0"
)

func Line() string {
	return Name + " " + Version
}

func Label() string {
	return Name + " (" + Version + ")"
}
