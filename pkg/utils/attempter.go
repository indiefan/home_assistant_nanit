package utils

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// AttemptContext - attempt context container
type AttemptContext interface {
	GracefulContext

	// GetTry - returns assigned number of tries
	GetTry() int
}

// PerseverenceOpts - options container for RunWithPerseverance
type PerseverenceOpts struct {
	// Cooldown - List of cooldown periods for failed attempts.
	// If execution fails more times than length of this array, last item is used.
	Cooldown []time.Duration

	// ResetThreshold - After this time failed attempts are counted as first failure
	ResetThreshold time.Duration

	// RunnerID - optional string name for the runner for debugging purposes
	RunnerID string
}

var lastPerseveranceRunnerID int32 = 0

// RunWithPerseverance - runs handler and tries it again if it fails
func RunWithPerseverance(handler func(AttemptContext), ctx GracefulContext, opts PerseverenceOpts) {
	try := 1
	timer := time.NewTimer(0)
	runnerID := fmt.Sprintf("runner%v", atomic.AddInt32(&lastPerseveranceRunnerID, 1))
	if opts.RunnerID != "" {
		runnerID = opts.RunnerID
	}

	sublog := log.With().Str("runner", runnerID).Logger()

	for {
		select {
		case <-ctx.Done():
			sublog.Trace().Msg("Perseverance run cancelled, there will be no further attempts")
			timer.Stop()
			return
		case timeScheduled := <-timer.C:
			hasBeenCancelled, err := ctx.RunAsChild(func(childGracefulCtx GracefulContext) {
				sublog.Trace().Int("try", try).Msg("Starting attempt")
				handler(newAttemptCtx(try, childGracefulCtx))
			}).Wait()

			if hasBeenCancelled {
				sublog.Trace().Msg("Perseverance run cancelled in the middle of execution, there will be no further attempts")
				timer.Stop()
				return
			}

			timeTaken := time.Since(timeScheduled)

			if err == nil {
				sublog.Trace().Msg("Attempt finished without an error")
				return
			}

			sublog.Trace().Err(err).Msg("Attempt finished with error")

			if opts.ResetThreshold > 0 && timeTaken > opts.ResetThreshold {
				sublog.Trace().Msgf("Previous attempt was %v ago, resetting tries", timeTaken)
				try = 1
				timer.Reset(0)
			} else {
				cooldown := opts.Cooldown[MinInt(try, len(opts.Cooldown))-1]
				try++

				if cooldown > timeTaken {
					waitDur := cooldown - timeTaken
					sublog.Trace().Str("cooldown", fmt.Sprintf("%v", cooldown)).Str("remaining", fmt.Sprintf("%v", waitDur)).Msg("Cooling down before the next attempt")
					timer.Reset(waitDur)
				} else {
					sublog.Trace().Msg("No cooldown necessary, we can perform next attempt")
					timer.Reset(0)
				}
			}
		}
	}
}

type attemptCtx struct {
	try         int
	gracefulCtx GracefulContext
}

func newAttemptCtx(try int, gracefulCtx GracefulContext) *attemptCtx {
	return &attemptCtx{try, gracefulCtx}
}

func (ctx *attemptCtx) GetTry() int           { return ctx.try }
func (ctx *attemptCtx) Done() <-chan struct{} { return ctx.gracefulCtx.Done() }
func (ctx *attemptCtx) Fail(err error)        { ctx.gracefulCtx.Fail(err) }
func (ctx *attemptCtx) RunAsChild(callback func(GracefulContext)) GracefulRunner {
	return ctx.gracefulCtx.RunAsChild(callback)
}
