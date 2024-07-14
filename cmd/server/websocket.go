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

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		app.logger.Error("dial:", err)
	}
	defer c.Close()

	app.wsConn = c

	done := make(chan struct{})

	// Goroutine to read messages from the WebSocket
	go func() {
		defer close(done)
		for {
			var msg Message
			// msgType, message, err := c.ReadMessage()
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Println("read:", err)
				return
			}
			// log.Printf("Received message: %s", message)
			// log.Printf("Message type: %d", msgType)
			app.logger.Info("message received", "type", msg.Type)
			switch msg.Type {
			case "status":
				// TODO: send group state using some function we need to create
				app.WriteGrouspStatusToClient()
			}
		}
	}()

	// Goroutine to send messages to the WebSocket
	go func() {
		for {
			// TODO: Use that function here as well, it will send the same message
			select {
			case <-done:
				return
			default:
				app.WriteGrouspStatusToClient()
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Wait for the read goroutine to complete
	<-done
	log.Println("WebSocket connection closed")
}

// updates app.groups and writes all hue groups to the server via websocket connection
func (app *application) WriteGrouspStatusToClient() {
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

// func (app *application) wsConnect() {
// 	fmt.Println("Connecting to WS")
// 	u := url.URL{Scheme: "ws", Host: app.config.remoteServerConfig.url, Path: "/ws"}
// 	app.logger.Info("connecting to ws", "url", u.String())
//
// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	if err != nil {
// 		app.logger.Error("dial:", err)
// 		return
// 	}
// 	defer c.Close()
//
// 	app.wsConn = c
//
// 	go func() {
// 		for {
// 			err := app.getGroups()
// 			if err != nil {
// 				app.logger.Error("getGroups:", err)
// 				return
// 			}
//
// 			msg := GroupsStateMessage{Type: "group_state", Data: *app.groups}
// 			err = c.WriteJSON(msg)
// 			if err != nil {
// 				app.logger.Error("write:", err)
// 				return
// 			}
//
// 			time.Sleep(5 * time.Second)
// 		}
// 	}()
//
// 	// Start a goroutine to read messages from the server
// 	go func() {
// 		for {
// 			var msg Message
// 			err := c.ReadJSON(&msg)
// 			if err != nil {
// 				app.logger.Error("read:", err)
// 				return
// 			}
// 			app.handleWSMessage(&msg)
// 		}
// 	}()
// }
//
// func (app application) handleWSMessage(msg *Message) {
// 	switch msg.Type {
// 	case "room_update":
// 		var roomUpdateMessage RoomUpdateMessage
// 		if err := mapToStruct(msg.Data, &roomUpdateMessage); err != nil {
// 			app.logger.Error("Error converting message data:", err)
// 			return
// 		}
// 		log.Println("Received room update message:", roomUpdateMessage)
// 	case "light_update":
// 		var lightUpdateMessage LightUpdateMessage
// 		if err := mapToStruct(msg.Data, &lightUpdateMessage); err != nil {
// 			app.logger.Error("Error converting message data:", err)
// 			return
// 		}
// 		app.logger.Error("Received light update message", "message", lightUpdateMessage)
// 	default:
// 		app.logger.Error("Unknown message type", "type", msg.Type)
// 	}
// }
//
// func mapToStruct(data interface{}, result interface{}) error {
// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}
// 	return json.Unmarshal(jsonData, result)
// }
