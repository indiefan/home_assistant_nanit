package client

import (
	"errors"
	"fmt"
	sync "sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sacOO7/gowebsocket"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
	"github.com/indiefan/home_assistant_nanit/pkg/session"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
	"google.golang.org/protobuf/proto"
)

type readyState struct {
	Context    utils.GracefulContext
	Connection *WebsocketConnection
}

// WebsocketConnectionHandler - handler of ready connection
type WebsocketConnectionHandler func(*WebsocketConnection, utils.GracefulContext)

// WebsocketConnectionManager - connection manager
type WebsocketConnectionManager struct {
	BabyUID          string
	CameraUID        string
	Session          *session.Session
	API              *NanitClient
	BabyStateManager *baby.StateManager

	mu               sync.RWMutex
	readyState       *readyState
	readySubscribers []WebsocketConnectionHandler
}

// NewWebsocketConnectionManager - constructor
func NewWebsocketConnectionManager(babyUID string, cameraUID string, session *session.Session, api *NanitClient, babyStateManager *baby.StateManager) *WebsocketConnectionManager {
	manager := &WebsocketConnectionManager{
		BabyUID:          babyUID,
		CameraUID:        cameraUID,
		Session:          session,
		API:              api,
		BabyStateManager: babyStateManager,
	}

	manager.WithReadyConnection(func(conn *WebsocketConnection, ctx utils.GracefulContext) {
		ticker := time.NewTicker(20 * time.Second)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				conn.SendMessage(&Message{
					Type: Message_Type(Message_KEEPALIVE).Enum(),
				})
			}
		}
	})

	return manager
}

// WithReadyConnection - registers handler which will be called as a go routine upon ready connection
func (manager *WebsocketConnectionManager) WithReadyConnection(handler WebsocketConnectionHandler) {
	manager.mu.Lock()
	readyState := manager.readyState
	manager.readySubscribers = append(manager.readySubscribers, handler)
	manager.mu.Unlock()

	if readyState != nil {
		log.Debug().Msg("Immediately notifying ready handler")
		notifyReadyHandler(handler, *readyState)
	}
}

// RunWithinContext - starts websocket connection attempt loop
func (manager *WebsocketConnectionManager) RunWithinContext(ctx utils.GracefulContext) {
	utils.RunWithPerseverance(manager.run, ctx, utils.PerseverenceOpts{
		RunnerID:       fmt.Sprintf("websocket-%v", manager.CameraUID),
		ResetThreshold: 2 * time.Second,
		Cooldown: []time.Duration{
			// 2 * time.Second,
			30 * time.Second,
			2 * time.Minute,
			15 * time.Minute,
			1 * time.Hour,
		},
	})
}

func (manager *WebsocketConnectionManager) run(attempt utils.AttemptContext) {
	// Reauthorize if it is not a first try or we assume we don't have a valid token
	manager.API.MaybeAuthorize(attempt.GetTry() > 1)

	// Remote
	url := fmt.Sprintf("wss://api.nanit.com/focus/cameras/%v/user_connect", manager.CameraUID)
	auth := fmt.Sprintf("Bearer %v", manager.Session.AuthToken)

	// Local
	// url := "wss://192.168.3.195:442"
	// auth := fmt.Sprintf("token %v", userCamToken)

	// -------

	var once sync.Once // Just because gowebsocket is buggy and can invoke OnDisconnect multiple times :-/

	socket := gowebsocket.New(url)
	socket.RequestHeader.Set("Authorization", auth)

	// Handle new connection
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Info().Str("url", url).Msg("Connected to websocket")

		go func() {
			conn := NewWebsocketConnection(&socket)
			readyState := readyState{attempt, conn}

			manager.mu.Lock()
			manager.readyState = &readyState
			subscribedHandlers := make([]WebsocketConnectionHandler, len(manager.readySubscribers))
			copy(subscribedHandlers, manager.readySubscribers)
			manager.mu.Unlock()

			manager.BabyStateManager.Update(manager.BabyUID, *baby.NewState().SetWebsocketAlive(true))

			log.Trace().Int("num_handlers", len(subscribedHandlers)).Msg("Notifying websocket ready handlers")

			for _, handler := range subscribedHandlers {
				notifyReadyHandler(handler, readyState)
			}
		}()
	}

	// Handle failed attempts for connection
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Error().Str("url", url).Err(err).Msg("Unable to establish websocket connection")
		attempt.Fail(err)
	}

	// Handle lost connection
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		once.Do(func() {
			manager.BabyStateManager.Update(manager.BabyUID, *baby.NewState().SetWebsocketAlive(false))

			if err != nil {
				log.Error().Err(err).Msg("Disconnected from server")
				attempt.Fail(err)
			} else {
				log.Warn().Msg("Disconnected from server")
				attempt.Fail(errors.New("Server closed the connection"))
			}
		})
	}

	socket.OnBinaryMessage = func(data []byte, _ gowebsocket.Socket) {
		m := &Message{}
		err := proto.Unmarshal(data, m)
		if err != nil {
			log.Error().Err(err).Bytes("rawdata", data).Msg("Received malformed binary message")
			return
		}

		log.Debug().Stringer("data", m).Msg("Received message")

		manager.mu.RLock()
		readyState := manager.readyState
		manager.mu.RUnlock()

		go readyState.Connection.handleMessage(m)
	}

	log.Trace().Msg("Connecting to websocket")
	socket.Connect()

	<-attempt.Done()

	if socket.IsConnected {
		log.Debug().Msg("Closing websocket")
		socket.Close()
	}
}

func notifyReadyHandler(handler WebsocketConnectionHandler, state readyState) {
	state.Context.RunAsChild(func(childCtx utils.GracefulContext) {
		handler(state.Connection, childCtx)
	})
}
