package version

const Maj = "0"
const Min = "1"
const Fix = "0"

var (
	// Version is the current version of teragrid
	// Must be a string because scripts like dist.sh read this file.
	Version = "0.1.0"

	// GitCommit is the current HEAD set using ldflags.
	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
