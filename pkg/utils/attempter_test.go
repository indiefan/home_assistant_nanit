package utils_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
)

func TestRunWithPerseverance(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC822}
	log.Logger = log.Output(consoleWriter)

	utils.RunWithGracefulCancel(func(ctx utils.GracefulContext) {
		utils.RunWithPerseverance(func(attempt utils.AttemptContext) {

			for {
				select {
				case <-time.After(100 * time.Millisecond):
					if attempt.GetTry() < 5 {
						attempt.Fail(fmt.Errorf("simulated failure %v", attempt.GetTry()))
					} else {
						return
					}
				case <-attempt.Done():
					return
				}
			}
		}, ctx, utils.PerseverenceOpts{
			ResetThreshold: 1 * time.Second,
			Cooldown:       []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond},
		})
	}).Wait()
}
