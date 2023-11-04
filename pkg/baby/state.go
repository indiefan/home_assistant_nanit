package baby

import (
	reflect "reflect"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

type StreamRequestState int32

const (
	StreamRequestState_NotRequested StreamRequestState = iota
	StreamRequestState_Requested
	StreamRequestState_RequestFailed
)

type StreamState int32

const (
	StreamState_Unknown StreamState = iota
	StreamState_Unhealthy
	StreamState_Alive
)

// State - struct holding information about state of a single baby
type State struct {
	StreamState        *StreamState        `internal:"true"`
	StreamRequestState *StreamRequestState `internal:"true"`
	IsWebsocketAlive   *bool               `internal:"true"`

	MotionTimestamp  *int32 // int32 is used to represent UTC timestamp
	SoundTimestamp   *int32 // int32 is used to represent UTC timestamp
	Temperature      *bool
	IsNight          *bool
	TemperatureMilli *int32
	HumidityMilli    *int32
}

// NewState - constructor
func NewState() *State {
	return &State{}
}

// Merge - Merges non-nil values of an argument to the state.
// Returns ptr to new state if changes
// Returns ptr to old state if not changed
func (state *State) Merge(stateUpdate *State) *State {
	newState := &State{}
	changed := false

	currReflect := reflect.ValueOf(state).Elem()
	newReflect := reflect.ValueOf(newState).Elem()
	patchReflect := reflect.ValueOf(stateUpdate).Elem()

	for i := 0; i < currReflect.NumField(); i++ {
		currField := currReflect.Field(i)
		newField := newReflect.Field(i)
		patchField := patchReflect.Field(i)

		if currField.Type().Kind() == reflect.Ptr {
			if !patchField.IsNil() && (currField.IsNil() || currField.Elem().Interface() != patchField.Elem().Interface()) {
				changed = true
				ptr := reflect.New(patchField.Type().Elem())
				ptr.Elem().Set(patchField.Elem())
				newField.Set(ptr)
			} else {
				newField.Set(currField)
			}
		}
	}

	if changed {
		return newState
	}

	return state
}

var upperCaseRX = regexp.MustCompile("[A-Z]+")

// AsMap - returns K/V map of non-nil properties
func (state *State) AsMap(includeInternal bool) map[string]interface{} {
	m := make(map[string]interface{})

	r := reflect.ValueOf(state).Elem()
	ts := reflect.TypeOf(*state)
	t := r.Type()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)

		if includeInternal || ts.Field(i).Tag.Get("internal") != "true" {

			if !f.IsNil() && f.Type().Kind() == reflect.Ptr {
				name := t.Field(i).Name
				var value interface{}

				if f.Type().Elem().Kind() == reflect.Int32 {
					value = f.Elem().Int()

					if strings.HasSuffix(name, "Milli") {
						name = strings.TrimSuffix(name, "Milli")
						value = float64(value.(int64)) / 1000
					}
				} else {
					value = f.Elem().Interface()
				}

				name = strings.ToLower(name[0:1]) + name[1:]
				name = upperCaseRX.ReplaceAllStringFunc(name, func(m string) string {
					return "_" + strings.ToLower(m)
				})

				m[name] = value
			}
		}
	}

	return m
}

// EnhanceLogEvent - appends non-nil properties to a log event
func (state *State) EnhanceLogEvent(e *zerolog.Event) *zerolog.Event {
	for key, value := range state.AsMap(true) {
		e.Interface(key, value)
	}

	return e
}

// SetTemperatureMilli - mutates field, returns itself
func (state *State) SetTemperatureMilli(value int32) *State {
	state.TemperatureMilli = &value
	return state
}

// GetTemperature - returns temperature as floating point
func (state *State) GetTemperature() float64 {
	if state.TemperatureMilli != nil {
		return float64(*state.TemperatureMilli) / 1000
	}

	return 0
}

// SetHumidityMilli - mutates field, returns itself
func (state *State) SetHumidityMilli(value int32) *State {
	state.HumidityMilli = &value
	return state
}

// GetHumidity - returns humidity as floating point
func (state *State) GetHumidity() float64 {
	if state.HumidityMilli != nil {
		return float64(*state.HumidityMilli) / 1000
	}

	return 0
}

// SetStreamRequestState - mutates field, returns itself
func (state *State) SetStreamRequestState(value StreamRequestState) *State {
	state.StreamRequestState = &value
	return state
}

// GetStreamRequestState - safely returns value
func (state *State) GetStreamRequestState() StreamRequestState {
	if state.StreamRequestState != nil {
		return *state.StreamRequestState
	}

	return StreamRequestState_NotRequested
}

// SetStreamState - mutates field, returns itself
func (state *State) SetStreamState(value StreamState) *State {
	state.StreamState = &value
	return state
}

// GetStreamState - safely returns value
func (state *State) GetStreamState() StreamState {
	if state.StreamState != nil {
		return *state.StreamState
	}

	return StreamState_Unknown
}

// SetIsNight - mutates field, returns itself
func (state *State) SetIsNight(value bool) *State {
	state.IsNight = &value
	return state
}

func (state *State) SetMotionTimestamp(value int32) *State {
	state.MotionTimestamp = &value
	return state
}

func (state *State) SetSoundTimestamp(value int32) *State {
	state.SoundTimestamp = &value
	return state
}

func (state *State) SetTemperature(value bool) *State {
	state.Temperature = &value
	return state
}

// GetIsWebsocketAlive - safely returns value
func (state *State) GetIsWebsocketAlive() bool {
	if state.StreamState != nil {
		return *state.IsWebsocketAlive
	}

	return false
}

// SetWebsocketAlive - mutates field, returns itself
func (state *State) SetWebsocketAlive(value bool) *State {
	state.IsWebsocketAlive = &value
	return state
}
