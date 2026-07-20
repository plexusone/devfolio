module github.com/plexusone/devfolio

go 1.26.2

// TODO: drop once omnidevx-core publishes a version tag containing the
// report/identity packages (currently only in the local working tree).
replace github.com/plexusone/omnidevx-core => ../omnidevx-core

require (
	github.com/google/go-github/v88 v88.0.0
	github.com/grokify/gogithub v0.13.0
	github.com/grokify/mogo v0.74.6
	github.com/grokify/structured-changelog v0.14.1
	github.com/plexusone/dashforge v0.3.0
	github.com/plexusone/omnidevx-core v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)
