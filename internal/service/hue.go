package service

import (
	"bufio"
	"fmt"
	"os"

	"github.com/amimof/huego"
)

type Hue struct {
	Bridge  *huego.Bridge
	Address string
}

func (h *Hue) DiscoverBridge(username string) error {
	// Call the discoverHueBridge function to find the IP address of the Hue bridge.
	if username == "" {
		return fmt.Errorf("Hue bridge username is required")
	}
	bridge, err := huego.Discover()
	if err != nil {
		return err
	}
	// Pause execution and ask the user to press the bridge's link button, and press enter to continue
	fmt.Printf("Discovered Hue Bridge IP: %s\n", bridge.Host)
	h.Address = bridge.Host
	fmt.Println("Please press the bridge's link button, then press enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	user, err := bridge.CreateUser(username)
	if err != nil {
		return err
	}

	bridge = bridge.Login(user)

	h.Bridge = bridge
	return nil
}

func (h *Hue) ConnectToBridge(ip, username string) error {
	bridge := huego.New(ip, username)
	h.Bridge = bridge
	h.Address = bridge.Host
	return nil
}
