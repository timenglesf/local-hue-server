package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/amimof/huego"
	"github.com/gorilla/websocket"
)

type StateMessage struct {
	BridgeID string                 `json:"bridge_id"`
	Data     map[string]interface{} `json:"data"`
}

type GroupsStateMessage struct {
	Type string `json:"type"`
	Data struct {
		Groups []huego.Group `json:"groups"`
	} `json:"data"`
}

type RoomUpdateMessage struct {
	RoomID string                 `json:"room_id"`
	Data   map[string]interface{} `json:"data"`
}

type LightUpdateMessage struct {
	LightID string                 `json:"light_id"`
	Data    map[string]interface{} `json:"data"`
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (app *application) wsConnect() {
	// Create a new WebSocket connection to the remote server.
	u := url.URL{Scheme: "ws", Host: app.config.remoteServerConfig.url, Path: "/ws"}
	for {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			app.logger.Error("dial:", err)
			time.Sleep(3 * time.Second)
			continue
		}
		defer c.Close()

		app.wsConn = c

		done := make(chan struct{})

		// Goroutine to read messages from the WebSocket
		go func() {
			defer close(done)
			for {
				var msg Message
				err := c.ReadJSON(&msg)
				if err != nil {
					log.Println("read:", err)
					return
				}

				app.logger.Info("message received", "type", msg.Type)
				switch msg.Type {
				case "status":
					app.WriteGroupsStatusToClient()
				case "update":
					fmt.Println("Update message received")
				}
			}
		}()

		// Goroutine to send messages to the WebSocket
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					app.WriteGroupsStatusToClient()
					time.Sleep(5 * time.Second)
				}
			}
		}()

		// Wait for the read goroutine to complete
		<-done
		log.Println("WebSocket connection closed")
	}
}

// updates app.groups and writes all hue groups to the server via websocket connection
func (app *application) WriteGroupsStatusToClient() {
	groups, err := app.hue.Bridge.GetGroups()
	if err != nil {
		app.logger.Error("Error getting groups", "error", err)
	} else {
		app.groups = &groups
	}
	fmt.Println("Groups have been fetched")
	msg := GroupsStateMessage{
		Type: "group_state",
		Data: struct {
			Groups []huego.Group `json:"groups"`
		}{
			Groups: *app.groups,
		},
	}

	err = app.wsConn.WriteJSON(msg)
	if err != nil {
		app.logger.Error("write:", err)
		return
	}
	fmt.Println("Sent message:", msg)
}
