package buildinfo

// BuildInfo cotains build information supplied at compile time.
type BuildInfo struct {
	Version    string // build version. e.g. v0.10.0
	CommitHash string // commit hash in vcs. e.g. git commit hash
}

var (
	// Those parameter can be supplied from compiler.
	// go build -ldflags "-X github.com/mzki/erago/infra/buildinfo.version=v0.1.2 -X github.com/mzki/erago/infra/buildinfo.commitHash=###"
	version    string = "dev"
	commitHash string = "none"
)

// Get returns BuildInfo filling with information supplied at compile time.
func Get() BuildInfo {
	return BuildInfo{
		Version:    version,
		CommitHash: commitHash,
	}
}
