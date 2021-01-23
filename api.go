package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amimof/huego"
)

// ConnectToHueAPI connects us to the Philips Hue REST API
func (l *Lightshow) ConnectToHueAPI() error {

	// If the bridge is already set, we probably don't want to create a new one
	if l.Bridge != nil {
		bridgeLogger := l.Logger.WithField("bridge-host", l.Bridge.Host)

		// If we can get capabilities from the bridge, we don't need to do anything
		capabilities, err := l.Bridge.GetCapabilities()
		if err == nil {
			bridgeLogger.WithField("capabilities", capabilities).Info("Connected to bridge")
			return nil
		}

		bridgeLogger.WithError(err).Warn("Failed to connect to bridge, retrying ConnectHueToAPI")
	}

	// Attempt to get the system hostname for registration
	hostname, err := os.Hostname()
	if err != nil {
		l.Logger.WithError(err).Error("Failed to get hostname for registration")
		return err
	}

	// Define the device type we will send to Hue API
	deviceType := fmt.Sprintf("lightshow#%s", hostname)

	if l.Config.Hue.Address == "" || l.Config.Hue.User == "" || l.Config.Hue.Secret == "" {
		// If one or more of the config values are empty, perform a registration

		// Attempt to discover bridge
		bridge, err := huego.Discover()
		if err != nil {
			l.Logger.WithError(err).Error("Failed to discover bridge")
			return err
		}

		bridgeLogger := l.Logger.WithField("bridge-host", bridge.Host)

		bridgeLogger.Info("Attempting to connect to bridge")

		// Setup a 30 second timeout context for connection
		connectCtx, connectCancel := context.WithTimeout(l.Context, time.Second*30)
		defer connectCancel()

		// Prepare to capture response if we successfully connect
		var whitelist *huego.Whitelist

		// Loop and attempt to connect
		for {

			// If the connect context has got an error, abort
			if connectCtx.Err() != nil {
				bridgeLogger.WithError(err).Error("Failed to connect to bridge")
				return connectCtx.Err()
			}

			time.Sleep(time.Second * 1)

			// Attempt to create the user on the bridge
			whitelist, err = bridge.CreateUserWithClientKeyContext(connectCtx, deviceType)
			if err != nil {
				if err.Error() != "ERROR 101 []: \"link button not pressed\"" {
					bridgeLogger.WithError(err).Error("Failed to create user on bridge")
					return err
				}
				bridgeLogger.WithError(err).Warn("Failed to create user on bridge")
			} else {
				bridgeLogger.Info("Created user")
				break
			}
		}

		// Update configuration values
		l.Config.Hue.Address = bridge.Host
		l.Config.Hue.User = whitelist.Username
		l.Config.Hue.Secret = whitelist.ClientKey

		// Save config
		err = l.SaveConfig()
		if err != nil {
			bridgeLogger.WithField("config", l.Config.Hue).Warn("Failed to save configuration file. Please manually save the config values")
		}
	}

	// Connect to bridge
	l.Bridge = huego.New(l.Config.Hue.Address, l.Config.Hue.User)

	// Loop through (if all went well, this should just succeed after calling GetCapabilities())
	return l.ConnectToHueAPI()
}
