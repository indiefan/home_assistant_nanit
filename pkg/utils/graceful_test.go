package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
)

func TestGracefulRunner(t *testing.T) {
	out := ""

	runner := utils.RunWithGracefulCancel(func(ctx utils.GracefulContext) {
		ctx.RunAsChild(func(childCtx utils.GracefulContext) {
			<-childCtx.Done()
			time.Sleep(500 * time.Millisecond)
			out = out + " sub_finished"
		})

		<-ctx.Done()
		time.Sleep(200 * time.Millisecond)
		out = out + " main_finished"
	})

	time.Sleep(100 * time.Millisecond)
	runner.Cancel()
	out = out + " after_cancel"

	assert.Equal(t, " main_finished sub_finished after_cancel", out)
}

func TestGracefulRunnerFinished(t *testing.T) {
	out := ""
	runner := utils.RunWithGracefulCancel(func(ctx utils.GracefulContext) {
		ctx.RunAsChild(func(childCtx utils.GracefulContext) {
			time.Sleep(300 * time.Millisecond)
			out = out + " sub_finished"
		})

		time.Sleep(200 * time.Millisecond)
		out = out + " main_finished"
	})

	_, err := runner.Wait()
	assert.NoError(t, err)
	assert.Equal(t, " main_finished sub_finished", out)
}

func TestGracefulRunnerFail(t *testing.T) {
	out := ""
	runner := utils.RunWithGracefulCancel(func(ctx utils.GracefulContext) {
		ctx.RunAsChild(func(childCtx utils.GracefulContext) {
			<-childCtx.Done()
			time.Sleep(200 * time.Millisecond)
			out = out + " sub_finished"
		})

		for {
			select {
			case <-time.After(200 * time.Millisecond):
				ctx.Fail(errors.New("simulated failure"))
			case <-ctx.Done():
				time.Sleep(100 * time.Millisecond)
				out = out + " main_finished"
				return
			}
		}
	})

	_, err := runner.Wait()
	assert.EqualError(t, err, "simulated failure")
	assert.Equal(t, " main_finished sub_finished", out)
}
