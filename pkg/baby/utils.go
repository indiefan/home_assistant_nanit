package baby

import (
	"regexp"

	"github.com/rs/zerolog/log"
)

var validUID = regexp.MustCompile(`^[a-z0-9_-]+$`)

// EnsureValidBabyUID - Checks that Baby UID does not contain any bad characters
// This is necessary because we use it as part of file paths
func EnsureValidBabyUID(babyUID string) {
	if !validUID.MatchString(babyUID) {
		log.Fatal().Str("uid", babyUID).Msg("Baby UID contains unsafe characters")
	}
}
