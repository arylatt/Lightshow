package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/amimof/huego"
)

// ConnectToHueAPI connects us to the Philips Hue REST API
func (l *Lightshow) ConnectToHueAPI() error {

	// If the bridge is already set, we probably don't want to create a new one
	if l.Bridge != nil {
		bridgeLogger := l.Logger.WithField("bridge-host", l.Bridge.Host)

		// If we can get capabilities from the bridge, we don't need to do anything
		config, err := l.Bridge.GetConfig()
		if err == nil {
			bridgeLogger.WithField("bridge-id", config.BridgeID).Info("Connected to bridge")
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

// ConfigureEntertainmentZone sets up the lights and entertainment zone
func (l *Lightshow) ConfigureEntertainmentZone() error {

	if l.Group != nil {
		l.Logger.Info("Entertainment zone has already been configured")
		return nil
	}

	// If we don't have the correct number of lights...
	if numLights := len(l.Config.Lights); numLights < 1 || numLights > 10 {
		err := fmt.Errorf("Expected 1-10 lights, got %d", numLights)
		l.Logger.WithError(err).Error("Failed to configure entertainment zone")
		return err
	}

	// Attempt to get the system hostname for enterainment zone name
	hostname, err := os.Hostname()
	if err != nil {
		l.Logger.WithError(err).Error("Failed to get hostname for entertainment zone creation")
		return err
	}

	groupName := fmt.Sprintf("lightshow-%s", hostname)

	// Prepare light data for Hue API
	locations := make(map[string][]float64, len(l.Config.Lights))
	lights := make([]string, len(l.Config.Lights))
	for i, light := range l.Config.Lights {
		lightID := strconv.Itoa(light.ID)
		lights[i] = lightID
		locations[lightID] = []float64{light.Coordinates.X, light.Coordinates.Y, 0}
	}

	// See if a group already exists
	groups, err := l.Bridge.GetGroupsContext(l.Context)
	if err != nil {
		l.Logger.WithError(err).Error("Faield to get groups")
		return err
	}

	var group *huego.Group
	for _, g := range groups {
		if g.Name == groupName {
			group = &g
			break
		}
	}

	updateGroup := huego.Group{
		Name:      groupName,
		Class:     "Other",
		Lights:    lights,
		Locations: locations,
	}

	createGroup := updateGroup
	createGroup.Type = "Entertainment"

	if group == nil {

		// Create group as it does not exist
		l.Logger.Info("Group does not exist. Creating...")

		success, err := l.Bridge.CreateGroupContext(l.Context, createGroup)
		if err != nil {
			l.Logger.WithError(err).Error("Failed to create group for entertainment zone")
			return err
		}

		groupID, _ := strconv.Atoi(success.Success["id"].(string))
		group = &huego.Group{ID: groupID}
	} else {

		// Update group as it exists
		l.Logger.Info("Group exists. Updating...")

		_, err = l.Bridge.UpdateGroupContext(l.Context, group.ID, updateGroup)
		if err != nil {
			l.Logger.WithError(err).Error("Failed to update group with lights and locations")
			return err
		}
	}

	// Fetch created/updated group
	l.Group, err = l.Bridge.GetGroupContext(l.Context, group.ID)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to get group we just created/updated")
		return err
	}

	l.Logger.Info("Created/updated group")
	return nil
}

// DeleteEntertainmentZone deletes the entertainment zone we created
func (l *Lightshow) DeleteEntertainmentZone() error {

	// Can't do anything if we don't have a group!
	if l.Group == nil {
		l.Logger.Warn("Called DeleteEntertainmentZone with no group configured")
		return nil
	}

	err := l.Bridge.DeleteGroup(l.Group.ID)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to delete group")
		return err
	}

	l.Group = nil
	l.Logger.Info("Deleted entertainment zone")
	return nil
}
