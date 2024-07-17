package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/amimof/huego"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type UpdateMessageData struct {
	Group      string `json:"group"`
	IsOn       bool   `json:"isOn"`
	Brightness *int   `json:"brightness,omitempty"` // HOW TO OMIT IF NOT SET
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
				fmt.Println(msg)
				switch msg.Type {
				case "status":
					app.WriteGroupsStatusToClient()
				case "update":
					err = app.UpdateGroup(msg)
					if err != nil {
						app.logger.Error("error updating group", "error", err)
					}
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

type GroupsStateMessage struct {
	Type string `json:"type"`
	Data struct {
		Groups []huego.Group `json:"groups"`
	} `json:"data"`
}

// updates app.groups and writes all hue groups to the server via websocket connection
func (app *application) WriteGroupsStatusToClient() {
	groups, err := app.hue.Bridge.GetGroups()
	if err != nil {
		app.logger.Error("error getting groups", "error", err)
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
}

func (app *application) UpdateGroup(msg Message) error {
	var data UpdateMessageData
	err := app.ConvertMessageData(msg.Data, &data)
	if err != nil {
		return err
	}

	for _, group := range *app.groups {
		if group.Name == data.Group {
			if data.IsOn {
				group.On()
			} else {
				group.Off()
			}
			if data.Brightness != nil {
				group.Bri(uint8(*data.Brightness))
			}
		}
	}

	return nil
}

// ConvertMessageData converts an interface{} to a struct using JSON marshalling and unmarshalling
func (app *application) ConvertMessageData(input interface{}, output interface{}) error {
	mData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}

	err = json.Unmarshal(mData, output)
	if err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}

	return nil
}
