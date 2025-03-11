package mqtt

import (
	"fmt"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
	"github.com/indiefan/home_assistant_nanit/pkg/client"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
	"github.com/rs/zerolog/log"
)

// Connection - MQTT context
type Connection struct {
	Opts         Opts
	StateManager *baby.StateManager
	client       MQTT.Client
}

// NewConnection - constructor
func NewConnection(opts Opts) *Connection {
	return &Connection{
		Opts: opts,
	}
}

// Run - runs the mqtt connection handler
func (conn *Connection) Run(manager *baby.StateManager, ctx utils.GracefulContext) {
	conn.StateManager = manager

	opts := MQTT.NewClientOptions()
	opts.AddBroker(conn.Opts.BrokerURL)
	opts.SetClientID(conn.Opts.TopicPrefix)
	opts.SetUsername(conn.Opts.Username)
	opts.SetPassword(conn.Opts.Password)
	opts.SetCleanSession(false)

	conn.client = MQTT.NewClient(opts)

	utils.RunWithPerseverance(func(attempt utils.AttemptContext) {
		runMqtt(conn, attempt)
	}, ctx, utils.PerseverenceOpts{
		RunnerID:       "mqtt",
		ResetThreshold: 2 * time.Second,
		Cooldown: []time.Duration{
			2 * time.Second,
			10 * time.Second,
			1 * time.Minute,
		},
	})
}

// // handleCommand - handles command messages from MQTT
// func (conn *Connection) handleLightCommand(mqttConn mqttClient.Client, msg mqttClient.Message) {
// 	// Extract baby UID and command from topic
// 	parts := strings.Split(msg.Topic(), "/")
// 	if len(parts) < 4 {
// 		log.Error().Str("topic", msg.Topic()).Msg("Invalid command topic format")
// 		return
// 	}

// 	babyUID := parts[2]
// 	command := parts[3]

// 	// Validate baby UID
// 	baby.EnsureValidBabyUID(babyUID)

// 	// Handle different commands
// 	switch command {
// 	case "light":
// 		enabled := string(msg.Payload()) == "on"
// 		log.Debug().
// 			Str("baby", babyUID).
// 			Bool("enabled", enabled).
// 			Str("payload", string(msg.Payload())).
// 			Msg("Received light command")

// 		// Get the baby info from the session
// 		for _, b := range conn.API.SessionStore.Session.Babies {
// 			if b.UID == babyUID {
// 				if ws := conn.wsProvider.GetWebsocket(babyUID); ws != nil {
// 					done := make(chan struct{})
// 					go func() {
// 						ws.WithReadyConnection(func(wsConn *client.WebsocketConnection, ctx utils.GracefulContext) {
// 							defer close(done)
// 							log.Debug().Msg("Inside WithReadyConnection callback - connection is ready")

// 							nightLight := client.Control_LIGHT_OFF
// 							if enabled {
// 								nightLight = client.Control_LIGHT_ON
// 							}

// 							// Create request with all required fields
// 							req := &client.Request{
// 								Type: client.RequestType_PUT_CONTROL.Enum(),
// 								Control: &client.Control{
// 									NightLight: &nightLight,
// 								},
// 							}

// 							// Send request and wait for response
// 							wsConn.SendRequest(client.RequestType_PUT_CONTROL, req)
// 						})
// 						log.Debug().Msg("WithReadyConnection returned")
// 					}()

// 					select {
// 					case <-done:
// 						log.Debug().Msg("Light command completed")
// 					case <-time.After(20 * time.Second):
// 						log.Error().Msg("Timeout waiting for WebSocket connection to be ready")
// 					}
// 				} else {
// 					log.Error().Str("baby", babyUID).Msg("Failed to get websocket connection")
// 				}
// 				return
// 			}
// 		}
// 		log.Error().Str("baby", babyUID).Msg("Baby not found in session")
// 	default:
// 		log.Warn().Str("command", command).Msg("Unknown command received")
// 	}
// }

func (conn *Connection) RegisterLightHandler(babyUID string, ws *client.WebsocketConnection) {
	commandTopic := fmt.Sprintf("%v/babies/%v/night_light/switch", conn.Opts.TopicPrefix, babyUID)
	log.Debug().
		Str("topic", commandTopic).
		Msg("Subscribing to command topic")

	lightCommandHandler := func(mqttConn MQTT.Client, msg MQTT.Message) {
		// Extract baby UID and command from topic
		parts := strings.Split(msg.Topic(), "/")
		if len(parts) < 4 {
			log.Error().Str("topic", msg.Topic()).Msg("Invalid command topic format")
			return
		}

		babyUID := parts[2]
		command := parts[4]

		// Validate baby UID
		baby.EnsureValidBabyUID(babyUID)

		// Handle different commands
		switch command {
		case "switch":
			enabled := string(msg.Payload()) == "true"
			log.Debug().
				Str("baby", babyUID).
				Bool("enabled", enabled).
				Str("payload", string(msg.Payload())).
				Msg("Received light command")

			nightLight := client.Control_LIGHT_OFF
			if enabled {
				nightLight = client.Control_LIGHT_ON
			}
			ws.SendRequest(client.RequestType_PUT_CONTROL, &client.Request{
				Control: &client.Control{
					NightLight: &nightLight,
				},
			})
		default:
			log.Warn().Str("command", command).Msg("Unknown command received")
		}
	}

	if token := conn.client.Subscribe(commandTopic, 0, lightCommandHandler); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Str("topic", commandTopic).Msg("Failed to subscribe to command topic")
		return
	}
}

func runMqtt(conn *Connection, attempt utils.AttemptContext) {

	if token := conn.client.Connect(); token.Wait() && token.Error() != nil {
		log.Error().Str("broker_url", conn.Opts.BrokerURL).Err(token.Error()).Msg("Unable to connect to MQTT broker")
		attempt.Fail(token.Error())
		return
	}

	log.Info().Str("broker_url", conn.Opts.BrokerURL).Msg("Successfully connected to MQTT broker")

	unsubscribe := conn.StateManager.Subscribe(func(babyUID string, state baby.State) {
		publish := func(key string, value interface{}) {
			topic := fmt.Sprintf("%v/babies/%v/%v", conn.Opts.TopicPrefix, babyUID, key)
			log.Trace().Str("topic", topic).Interface("value", value).Msg("MQTT publish")

			token := conn.client.Publish(topic, 0, false, fmt.Sprintf("%v", value))
			if token.Wait(); token.Error() != nil {
				log.Error().Err(token.Error()).Msgf("Unable to publish %v update", key)
			}
		}

		for key, value := range state.AsMap(false) {
			publish(key, value)
		}

		if state.StreamState != nil && *state.StreamState != baby.StreamState_Unknown {
			publish("is_stream_alive", *state.StreamState == baby.StreamState_Alive)
		}
	})

	// Wait until interrupt signal is received
	<-attempt.Done()

	log.Debug().Msg("Closing MQTT connection on interrupt")
	unsubscribe()
	conn.client.Disconnect(250)
}
