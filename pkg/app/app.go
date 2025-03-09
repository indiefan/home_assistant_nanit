package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/indiefan/home_assistant_nanit/pkg/baby"
	"github.com/indiefan/home_assistant_nanit/pkg/client"
	"github.com/indiefan/home_assistant_nanit/pkg/message"
	"github.com/indiefan/home_assistant_nanit/pkg/mqtt"
	"github.com/indiefan/home_assistant_nanit/pkg/rtmpserver"
	"github.com/indiefan/home_assistant_nanit/pkg/session"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
)

// App - application container
type App struct {
	Opts             Opts
	SessionStore     *session.Store
	BabyStateManager *baby.StateManager
	RestClient       *client.NanitClient
	MQTTConnection   *mqtt.Connection
	websockets       map[string]*client.WebsocketConnectionManager
}

// NewApp - constructor
func NewApp(opts Opts) *App {
	sessionStore := session.InitSessionStore(opts.SessionFile)

	instance := &App{
		Opts:             opts,
		BabyStateManager: baby.NewStateManager(),
		SessionStore:     sessionStore,
		RestClient: &client.NanitClient{
			Email:        opts.NanitCredentials.Email,
			Password:     opts.NanitCredentials.Password,
			RefreshToken: opts.NanitCredentials.RefreshToken,
			SessionStore: sessionStore,
		},
	}

	if opts.MQTT != nil {
		instance.MQTTConnection = mqtt.NewConnection(*opts.MQTT)
	}

	return instance
}

// Run - application main loop
func (app *App) Run(ctx utils.GracefulContext) {
	// Reauthorize if we don't have a token or we assume it is invalid
	app.RestClient.MaybeAuthorize(false)

	// Fetches babies info if they are not present in session
	app.RestClient.EnsureBabies()

	// RTMP
	if app.Opts.RTMP != nil {
		go rtmpserver.StartRTMPServer(app.Opts.RTMP.ListenAddr, app.BabyStateManager)
	}

	// MQTT
	if app.MQTTConnection != nil {
		ctx.RunAsChild(func(childCtx utils.GracefulContext) {
			app.MQTTConnection.Run(app.BabyStateManager, childCtx)
		})
	}

	// Start reading the data from the stream
	for _, babyInfo := range app.SessionStore.Session.Babies {
		_babyInfo := babyInfo
		ctx.RunAsChild(func(childCtx utils.GracefulContext) {
			app.handleBaby(_babyInfo, childCtx)
		})
	}

	// Start serving content over HTTP
	if app.Opts.HTTPEnabled {
		go serve(app.SessionStore.Session.Babies, app.Opts.DataDirectories)
	}

	<-ctx.Done()
}

func (app *App) handleBaby(baby baby.Baby, ctx utils.GracefulContext) {
	if app.Opts.RTMP != nil || app.MQTTConnection != nil {
		// Websocket connection
		ws := client.NewWebsocketConnectionManager(baby.UID, baby.CameraUID, app.SessionStore.Session, app.RestClient, app.BabyStateManager)

		ws.WithReadyConnection(func(conn *client.WebsocketConnection, childCtx utils.GracefulContext) {
			app.runWebsocket(baby.UID, conn, childCtx)
		})

		if app.Opts.EventPolling.Enabled {
			go app.pollMessages(baby.UID, app.BabyStateManager)
		}

		ctx.RunAsChild(func(childCtx utils.GracefulContext) {
			ws.RunWithinContext(childCtx)
		})
	}

	<-ctx.Done()
}

func (app *App) pollMessages(babyUID string, babyStateManager *baby.StateManager) {
	newMessages := app.RestClient.FetchNewMessages(babyUID, app.Opts.EventPolling.MessageTimeout)

	for {
		for _, msg := range newMessages {
			switch msg.Type {
			case message.SoundEventMessageType:
				go babyStateManager.NotifySoundSubscribers(babyUID, time.Time(msg.Time))
			case message.MotionEventMessageType:
				go babyStateManager.NotifyMotionSubscribers(babyUID, time.Time(msg.Time))
			}
		}

		// wait for the specified interval
		time.Sleep(app.Opts.EventPolling.PollingInterval)
	}
}

func (app *App) runWebsocket(babyUID string, conn *client.WebsocketConnection, childCtx utils.GracefulContext) {
	// Reading sensor data
	conn.RegisterMessageHandler(func(m *client.Message, conn *client.WebsocketConnection) {
		// Sensor request initiated by us on start (or some other client, we don't care)
		if *m.Type == client.Message_RESPONSE && m.Response != nil {
			if *m.Response.RequestType == client.RequestType_GET_SENSOR_DATA && len(m.Response.SensorData) > 0 {
				processSensorData(babyUID, m.Response.SensorData, app.BabyStateManager)
			}
		} else

		// Communication initiated from a cam
		// Note: it sends the updates periodically on its own + whenever some significant change occurs
		if *m.Type == client.Message_REQUEST && m.Request != nil {
			if *m.Request.Type == client.RequestType_PUT_SENSOR_DATA && len(m.Request.SensorData_) > 0 {
				processSensorData(babyUID, m.Request.SensorData_, app.BabyStateManager)
			} else if *m.Request.Type == client.RequestType_PUT_CONTROL && m.Request.Control != nil {
				processLight(babyUID, m.Request.Control, app.BabyStateManager)
			}
		}
	})

	app.MQTTConnection.RegisterLightHandler(babyUID, conn)

	// Ask for sensor data (initial request)
	conn.SendRequest(client.RequestType_GET_SENSOR_DATA, &client.Request{
		GetSensorData: &client.GetSensorData{
			All: utils.ConstRefBool(true),
		},
	})

	// Ask for status
	// conn.SendRequest(client.RequestType_GET_STATUS, &client.Request{
	// 	GetStatus_: &client.GetStatus{
	// 		All: utils.ConstRefBool(true),
	// 	},
	// })

	// Ask for logs
	// conn.SendRequest(client.RequestType_GET_LOGS, &client.Request{
	// 	GetLogs: &client.GetLogs{
	// 		Url: utils.ConstRefStr("http://192.168.3.234:8080/log"),
	// 	},
	// })

	var cleanup func()

	// Local streaming
	if app.Opts.RTMP != nil {
		initializeLocalStreaming := func() {
			requestLocalStreaming(babyUID, app.getLocalStreamURL(babyUID), client.Streaming_STARTED, conn, app.BabyStateManager)
		}

		// Watch for stream liveness change
		unsubscribe := app.BabyStateManager.Subscribe(func(updatedBabyUID string, stateUpdate baby.State) {
			// Do another streaming request if stream just turned unhealthy
			if updatedBabyUID == babyUID && stateUpdate.StreamState != nil && *stateUpdate.StreamState == baby.StreamState_Unhealthy {
				// Prevent duplicate request if we already received failure
				if app.BabyStateManager.GetBabyState(babyUID).GetStreamRequestState() != baby.StreamRequestState_RequestFailed {
					go initializeLocalStreaming()
				}
			}
		})

		cleanup = func() {
			// Stop listening for stream liveness change
			unsubscribe()

			// Stop local streaming
			state := app.BabyStateManager.GetBabyState(babyUID)
			if state.GetIsWebsocketAlive() && state.GetStreamState() == baby.StreamState_Alive {
				requestLocalStreaming(babyUID, app.getLocalStreamURL(babyUID), client.Streaming_STOPPED, conn, app.BabyStateManager)
			}
		}

		// Initialize local streaming upon connection if we know that the stream is not alive
		babyState := app.BabyStateManager.GetBabyState(babyUID)
		if babyState.GetStreamState() != baby.StreamState_Alive {
			if babyState.GetStreamRequestState() != baby.StreamRequestState_Requested || babyState.GetStreamState() == baby.StreamState_Unhealthy {
				go initializeLocalStreaming()
			}
		}
	}

	<-childCtx.Done()
	if cleanup != nil {
		cleanup()
	}
}

func (app *App) getRemoteStreamURL(babyUID string) string {
	return fmt.Sprintf("rtmps://media-secured.nanit.com/nanit/%v.%v", babyUID, app.SessionStore.Session.AuthToken)
}

func (app *App) getLocalStreamURL(babyUID string) string {
	if app.Opts.RTMP != nil {
		tpl := "rtmp://{publicAddr}/local/{babyUid}"
		return strings.NewReplacer("{publicAddr}", app.Opts.RTMP.PublicAddr, "{babyUid}", babyUID).Replace(tpl)
	}

	return ""
}
