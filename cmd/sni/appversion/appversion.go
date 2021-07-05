package appversion

var (
	Version string
	Commit  string
	Date    string
	BuiltBy string
)

func Init(
	version string,
	commit string,
	date string,
	builtBy string,
) {
	Version = version
	Commit = commit
	Date = date
	BuiltBy = builtBy
}
