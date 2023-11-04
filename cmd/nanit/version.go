package main

import "github.com/rs/zerolog/log"

// GitCommit - Injected on CI (from CI_COMMIT_SHORT_SHA)
var GitCommit string

func logAppVersion() {
	initMsg := log.Info()
	if GitCommit != "" {
		initMsg.Str("gitversion", GitCommit)
	}

	initMsg.Msg("Application started")
}
