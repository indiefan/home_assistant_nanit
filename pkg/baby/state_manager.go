package baby

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// StateManager - state manager context
type StateManager struct {
	babiesByUID      map[string]State
	subscribers      map[*chan bool]func(babyUID string, state State)
	stateMutex       sync.RWMutex
	subscribersMutex sync.RWMutex
}

// NewStateManager - state manager constructor
func NewStateManager() *StateManager {
	return &StateManager{
		babiesByUID: make(map[string]State),
		subscribers: make(map[*chan bool]func(babyUID string, state State)),
	}
}

// Update - updates baby info in thread safe manner
func (manager *StateManager) Update(babyUID string, stateUpdate State) {
	var updatedState *State

	manager.stateMutex.Lock()
	defer manager.stateMutex.Unlock()

	if babyState, ok := manager.babiesByUID[babyUID]; ok {
		updatedState = babyState.Merge(&stateUpdate)
		if updatedState == &babyState {
			return
		}
	} else {
		updatedState = NewState().Merge(&stateUpdate)
	}

	manager.babiesByUID[babyUID] = *updatedState
	stateUpdate.EnhanceLogEvent(log.Debug().Str("baby_uid", babyUID)).Msg("Baby state updated")

	go manager.notifySubscribers(babyUID, stateUpdate)
}

// Subscribe - registers function to be called on every update
// Returns unsubscribe function
func (manager *StateManager) Subscribe(callback func(babyUID string, state State)) func() {
	unsubscribeC := make(chan bool, 1)

	manager.subscribersMutex.Lock()
	manager.subscribers[&unsubscribeC] = callback
	manager.subscribersMutex.Unlock()

	manager.stateMutex.RLock()
	for babyUID, babyState := range manager.babiesByUID {
		go callback(babyUID, babyState)
	}

	manager.stateMutex.RUnlock()

	return func() {
		manager.subscribersMutex.Lock()
		delete(manager.subscribers, &unsubscribeC)
		manager.subscribersMutex.Unlock()
	}
}

// GetBabyState - returns current state of a baby
func (manager *StateManager) GetBabyState(babyUID string) *State {
	manager.stateMutex.RLock()
	babyState := manager.babiesByUID[babyUID]
	manager.stateMutex.RUnlock()

	return &babyState
}

func (manager *StateManager) NotifyMotionSubscribers(babyUID string, time time.Time) {
	timestamp := new(int32)
	*timestamp = int32(time.Unix())
	var state = State{MotionTimestamp: timestamp}

	manager.notifySubscribers(babyUID, state)
}

func (manager *StateManager) NotifySoundSubscribers(babyUID string, time time.Time) {
	timestamp := new(int32)
	*timestamp = int32(time.Unix())
	var state = State{SoundTimestamp: timestamp}

	manager.notifySubscribers(babyUID, state)
}

func (manager *StateManager) notifySubscribers(babyUID string, state State) {
	manager.subscribersMutex.RLock()

	for _, callback := range manager.subscribers {
		go callback(babyUID, state)
	}

	manager.subscribersMutex.RUnlock()
}
