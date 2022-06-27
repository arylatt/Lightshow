package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/amimof/huego"
	"github.com/arylatt/lightshow"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mapCmd = &cobra.Command{
	Use:   "play-map",
	Short: "Play a specific map file.",
	RunE:  playMap,
}

func init() {
	mapCmd.Flags().StringP("map-file", "m", "", "Path to a Map file")
	mapCmd.MarkFlagFilename("map-file", ".json")
	viper.BindPFlag("mapfile", mapCmd.Flags().Lookup("map-file"))

	rootCmd.AddCommand(mapCmd)
}

func playMap(cmd *cobra.Command, args []string) error {
	viper.AllSettings()
	show := &lightshow.Lightshow{}

	mapBytes, err := ioutil.ReadFile(viper.GetString("mapfile"))
	if err != nil {
		return fmt.Errorf("error reading map file: %w", err)
	}

	showMap := &lightshow.Map{}
	err = json.Unmarshal(mapBytes, &showMap)
	if err != nil {
		return fmt.Errorf("error decoding map file: %w", err)
	}

	if err = showMap.Validate(); err != nil {
		return fmt.Errorf("error validating map file: %w", err)
	}

	bridge := huego.New(viper.GetString("host"), viper.GetString("user"))
	userID, err := lightshow.GetApplicationID(viper.GetString("host"), viper.GetString("user"))
	if err != nil {
		return fmt.Errorf("error connecting to bridge: %w", err)
	}

	group, err := bridge.GetGroup(viper.GetInt("group"))
	if err != nil {
		return fmt.Errorf("error looking up entertainment group: %w", err)
	}

	if err = show.OpenDTLSConnection(cmd.Context(), group, viper.GetString("host"), userID, viper.GetString("key"), viper.GetInt("port")); err != nil {
		return fmt.Errorf("error establishing Hue DTLS connection: %w", err)
	}

	t := time.Now()
	err = show.RunMap(cmd.Context(), *showMap)
	fmt.Printf("Total elapsed time for map: %f\r\n", time.Since(t).Seconds())
	return err
}
