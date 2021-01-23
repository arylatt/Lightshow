package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/sirupsen/logrus"
)

// StartDTLSStreaming runs the routine to continously
func (l *Lightshow) StartDTLSStreaming() {

	// Can't start streaming if there is no group...
	if l.Group == nil {
		l.Logger.Error("Can't start DTLS streaming if there is no group")
		l.ContextCancelFunc()
		return
	}

	// Initiate streaming
	err := l.Group.EnableStreamingContext(l.Context)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to start streaming")
		return
	}
	l.Logger.Info("Enabled streaming for entertainment zone")

	// Setup the address of the DTLS server
	dtlsAddress := &net.UDPAddr{
		IP:   net.ParseIP(l.Config.Hue.Address),
		Port: 2100,
	}

	// Setup the authn data for DTLS connection
	dtlsConfig := &dtls.Config{
		PSKIdentityHint: []byte(l.Bridge.User),
		CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_GCM_SHA256},
		PSK: func(hint []byte) ([]byte, error) {
			return hex.DecodeString(l.Config.Hue.Secret)
		},
	}

	// Prepare context for connecting to DTLS server
	dtlsCtx, dtlsCtxCancel := context.WithTimeout(l.Context, time.Second*5)
	defer dtlsCtxCancel()

	l.DTLSConn, err = dtls.DialWithContext(dtlsCtx, "udp", dtlsAddress, dtlsConfig)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to connect to DTLS server")
		return
	}

	go l.DTLSMessageLoop()
}

// DTLSMessageLoop loops around and sends whatever messages we need to send
func (l *Lightshow) DTLSMessageLoop() {

	// No point sending anything if we don't have a connection
	if l.DTLSConn == nil {
		l.Logger.Error("Tried to call DTLSMessageLoop without a DTLS connection")
		l.ContextCancelFunc()
		return
	}

	messageHeader := []byte("HueStream")
	messageHeader = append(messageHeader, []byte{
		0x01, 0x00, // Version number
		0x00,       // Sequence number (ignored)
		0x00, 0x00, // Reserved
		0x01, // Color mode XY+Brightness
		0x00, // Reserved
	}...)

	sleepDuration := time.Millisecond * time.Duration((float64(time.Second/time.Millisecond) / l.Config.Hue.FrequencyHz))

	l.Logger.WithFields(logrus.Fields{
		"frequency": fmt.Sprintf("%f Hz", l.Config.Hue.FrequencyHz),
	}).Info("Starting DTLS message loop")

	for {
		if l.Context.Err() != nil {
			l.Logger.Info("Received context cancellation, aborting DTLSMessageLoop")
			return
		}

		_, err := l.DTLSConn.Write(append(messageHeader, l.MessageBytes...))
		if err != nil {
			l.Logger.WithError(err).Error("Received error writing to DTLS connection")
			l.ContextCancelFunc()
			return
		}

		if l.Context.Err() == nil {
			time.Sleep(sleepDuration)
		}
	}
}
