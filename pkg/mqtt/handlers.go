package mqtt

import (
    "github.com/indiefan/home_assistant_nanit/pkg/baby"
    "github.com/indiefan/home_assistant_nanit/pkg/client"
)

// RegisterLightHandler registers a handler for light control messages
func RegisterLightHandler(babyUID string, conn *client.WebsocketConnection, stateManager *baby.StateManager) {
    stateManager.Subscribe(func(updatedBabyUID string, state baby.State) {
        if updatedBabyUID == babyUID && state.NightLight != nil {
            requestLightControl(*state.NightLight, conn)
        }
    })
}

func requestLightControl(enabled bool, conn *client.WebsocketConnection) {
    nightLight := client.Control_LIGHT_OFF
    if enabled {
        nightLight = client.Control_LIGHT_ON
    }

    conn.SendRequest(client.RequestType_PUT_CONTROL, &client.Request{
        Control: &client.Control{
            NightLight: &nightLight,
        },
    })
} 