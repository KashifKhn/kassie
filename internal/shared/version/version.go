package version

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func Get() string {
	return Version
}

func GetFull() string {
	return Version + " (" + Commit + ") built on " + BuildDate
}
