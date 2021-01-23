package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/amimof/huego"
	"github.com/pion/dtls/v2"
	"github.com/spf13/viper"
)

type hueconfig struct {
	Address string
	User    string
	Secret  string
}

type config struct {
	Huecfg hueconfig
}

var appcfg config
var bridge *huego.Bridge

func (c *config) connectToHue() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	if c.Huecfg.Address == "" || c.Huecfg.User == "" || c.Huecfg.Secret == "" {
		bridge, err = huego.Discover()
		if err != nil {
			return err
		}

		fmt.Println(fmt.Sprintf("Discovered bridge '%s', press link button and then press any key to continue", bridge.Host))

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		whitelist, err := bridge.CreateUserWithClientKey(fmt.Sprintf("hueapp-golang#%s", hostname))
		if err != nil {
			return err
		}

		c.Huecfg.Address = bridge.Host
		c.Huecfg.User = whitelist.Username
		c.Huecfg.Secret = whitelist.ClientKey

		bridge = bridge.Login(whitelist.Username)
	} else {
		bridge = huego.New(c.Huecfg.Address, c.Huecfg.User)
	}

	out, err := json.Marshal(c)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(out)
	err = viper.MergeConfig(reader)
	if err != nil {
		return err
	}

	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Failed to read in config: %s\n", err)
		return
	}

	err = viper.Unmarshal(&appcfg)
	if err != nil {
		fmt.Printf("Failed to unmarshal config: %s\n", err)
		return
	}

	err = appcfg.connectToHue()
	if err != nil {
		fmt.Printf("Failed to connect: %s\n", err)
		return
	}

	signals := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		sig := <-signals
		fmt.Printf("\n%s\n", sig)
		cancel()
	}()

	group := &huego.Group{
		Name:   "andrew-hueapp-test",
		Lights: []string{"5", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
		Type:   "Entertainment",
		Class:  "Other",
	}

	resp, err := bridge.CreateGroup(*group)
	if err != nil {
		fmt.Printf("Failed to create group: %s\n", err)
		return
	}

	defer func() {
		err := bridge.DeleteGroup(group.ID)
		if err != nil {
			fmt.Printf("Error deleting group: %s\n", err)
		}
	}()

	group.ID, err = strconv.Atoi(resp.Success["id"].(string))
	if err != nil {
		fmt.Printf("Failed to parse group ID: %s\n", err)
		return
	}

	group, err = bridge.GetGroup(group.ID)
	if err != nil {
		fmt.Printf("Failed to get group: %s\n", err)
		return
	}

	// group.Alert("select")

	err = group.EnableStreaming()
	if err != nil {
		fmt.Printf("Failed to enable streaming: %s\n", err)
		return
	}

	defer func() {
		err := group.DisableStreaming()
		if err != nil {
			fmt.Printf("Error disabling streaming: %s\n", err)
		}
	}()

	dtlsAddress := &net.UDPAddr{
		IP:   net.ParseIP(appcfg.Huecfg.Address),
		Port: 2100,
	}

	dtlsConfig := &dtls.Config{
		PSKIdentityHint: []byte(bridge.User),
		CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_GCM_SHA256},
		PSK: func(hint []byte) ([]byte, error) {
			fmt.Printf("Server's hint: %s \n", hint)
			return hex.DecodeString(appcfg.Huecfg.Secret)
		},
	}

	dialCtx, dialCancel := context.WithTimeout(ctx, time.Second*30)
	defer dialCancel()

	dtlsConnection, err := dtls.DialWithContext(dialCtx, "udp", dtlsAddress, dtlsConfig)
	if err != nil {
		fmt.Printf("Failed to establish DTLS connection: %s\n", err)
		return
	}

	defer dtlsConnection.Close()

	header := []byte("HueStream")

	messageOn := append(header, []byte{
		0x01, 0x00,
		0x00,
		0x00, 0x00,
		0x01,
		0x00,
		0x00, 0x00, 0x05,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x07,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x08,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x09,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x0a,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x0b,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x0c,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x0d,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x0e,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x0f,
		0x55, 0xff, 0x55, 0xff, 0xff, 0xff,
	}...)

	messageOff := append(header, []byte{
		0x01, 0x00,
		0x00,
		0x00, 0x00,
		0x01,
		0x00,
		0x00, 0x00, 0x05,
		0x55, 0xff, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x07,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x08,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x09,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0a,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0b,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0c,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0d,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0e,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0f,
		0x55, 0x00, 0x55, 0x00, 0x00, 0x00,
	}...)

	flashDuration := time.Millisecond * 100

	for {

		if ctx.Err() != nil {
			return
		}

		_, err := write(dtlsConnection, messageOn)
		if err != nil {
			return
		}

		time.Sleep(flashDuration)

		_, err = write(dtlsConnection, messageOff)
		if err != nil {
			return
		}

		time.Sleep((time.Millisecond * 341) - flashDuration)
	}

}

func write(connection *dtls.Conn, msg []byte) (int, error) {
	for _, m := range msg {
		fmt.Printf("0x%s, ", hex.EncodeToString([]byte{m}))
	}
	fmt.Printf("\n")

	len, err := connection.Write(msg)
	if err != nil {
		fmt.Printf("Error writing: %s\n", err)
		return 0, err
	}

	return len, err
}
