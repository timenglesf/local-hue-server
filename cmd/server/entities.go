package main

import (
	"fmt"

	"github.com/amimof/huego"
)

type Message struct {
	Data interface{} `json:"data"`
	Type messageType `json:"type"`
}

type UpdateMessageData struct {
	Brightness *int   `json:"brightness,omitempty"` // HOW TO OMIT IF NOT SET
	Group      string `json:"group"`
	IsOn       bool   `json:"isOn"`
}

type messageType string

var (
	statusMessage messageType = "status"
	updateMessage messageType = "update"
)

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
