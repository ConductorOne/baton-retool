package connector

var (
	ownLevel                = "own"
	writeLevel              = "write"
	readLevel               = "read"
	noneLevel               = "none"
	accessLevels            = []string{readLevel, writeLevel, ownLevel}
	resourceAccessLevels    = []string{readLevel, writeLevel}
	accessLevelDisplayNames = map[string]string{
		readLevel:  "use",
		writeLevel: "edit",
		ownLevel:   "own",
	}
)
