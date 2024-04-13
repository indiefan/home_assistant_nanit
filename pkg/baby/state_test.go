package baby_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
)

func TestStateAsMap(t *testing.T) {
	s := baby.State{}
	s.SetTemperatureMilli(1000)
	s.SetIsNight(true)
	s.SetStreamRequestState(baby.StreamRequestState_Requested)

	m := s.AsMap(false)

	assert.Equal(t, 1.0, m["temperature"], "The two words should be the same.")
	assert.Equal(t, true, m["is_night"], "The two words should be the same.")
	assert.NotContains(t, m, "is_stream_requested", "Should not contain internal fields")
}

func TestStateMergeSame(t *testing.T) {
	s1 := &baby.State{}
	s1.SetTemperatureMilli(10)

	s2 := &baby.State{}
	s2.SetTemperatureMilli(10)

	s3 := s1.Merge(s2)
	assert.Same(t, s1, s3)
}

func TestStateMergeDifferent(t *testing.T) {
	s1 := &baby.State{}
	s1.SetTemperatureMilli(10_000)
	s1.SetStreamState(baby.StreamState_Alive)

	s2 := &baby.State{}
	s2.SetTemperatureMilli(11_000)
	s2.SetHumidityMilli(20_000)
	s2.SetStreamState(baby.StreamState_Alive)

	s3 := s1.Merge(s2)
	assert.NotSame(t, s1, s3)
	assert.NotSame(t, s2, s3)
	assert.Equal(t, 10.0, s1.GetTemperature())

	assert.Equal(t, 11.0, s3.GetTemperature())
	assert.Equal(t, 20.0, s3.GetHumidity())
	assert.Equal(t, baby.StreamState_Alive, s3.GetStreamState())
}
