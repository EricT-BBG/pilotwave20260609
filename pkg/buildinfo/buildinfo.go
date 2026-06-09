package buildinfo

const unknown = "unknown"

var (
	Version   = "1.3.1"
	Commit    = unknown
	BuildTime = unknown
)

func Values() (string, string, string) {
	return Version, Commit, BuildTime
}
