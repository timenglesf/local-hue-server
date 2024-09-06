package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func (app *application) wsConnect() {
	for {
		// Create a new WebSocket connection to the remote server.
		u := url.URL{Scheme: "ws", Host: app.config.remoteServerConfig.url, Path: "/ws"}

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			app.logger.Error("dial:", err)
			time.Sleep(3 * time.Second)
		}

		app.wsConn = c

		go app.writePump()
		go app.readPump()
	}
}

func (app *application) readPump() {
	defer func() {
		err := app.wsConn.Close()
		if err != nil {
			app.logger.Error("error closing ws connection", "error", err)
		}
	}()
	for {
		var msg Message
		err := app.wsConn.ReadJSON(&msg)
		if err != nil {
			app.logger.Error("error reading msg:", "error", err, "msg", msg)
			return
		}
		app.logger.Info("message received", "type", msg.Type)

		switch msg.Type {
		case statusMessage:
			app.WriteGroupsStatusToClient()
		case updateMessage:
			err = app.UpdateGroup(msg)
			if err != nil {
				app.logger.Error("error updating group", "error", err)
			}
		}
	}
}

func (app *application) writePump() {
	for {
		app.WriteGroupsStatusToClient()
		time.Sleep(5 * time.Second)
	}
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
