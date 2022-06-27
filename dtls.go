package lightshow

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	"github.com/amimof/huego"
	"github.com/pion/dtls/v2"
)

var (
	// HueDTLSProtocolName is the byte-encoded version of HueStream as defined in API docs
	HueDTLSProtocolName = []byte("HueStream")

	// HueDTLSHeader is the required header data to send before the light data
	HueDTLSHeader = append(HueDTLSProtocolName, []byte{
		0x01, 0x00, // Version number
		0x00,       // Sequence number (ignored)
		0x00, 0x00, // Reserved
		0x01, // Color mode XY+Brightness
		0x00, // Reserved
	}...)

	// ErrDTLSConnNil is returned if write is called on an empty connection
	ErrDTLSConnNil = errors.New("dtls connection is nil")

	// ErrDTLSOpenNilGroup indicates when a DTLS connection was not established because no Hue group was specified
	ErrDTLSOpenNilGroup = errors.New("attempted to open dtls connection with nil hue group")
)

// OpenDTLSConnection will mark a light group as enabled for streaming and then open a DTLS connection for streaming data
func (l *Lightshow) OpenDTLSConnection(ctx context.Context, group *huego.Group, address, user, secret string, port int) error {
	errFmt := "error opening dtls connection: %w"

	if group == nil {
		return fmt.Errorf(errFmt, ErrDTLSOpenNilGroup)
	}

	err := group.EnableStreamingContext(ctx)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	dtlsAddress := &net.UDPAddr{
		IP:   net.ParseIP(address),
		Port: port,
	}

	dtlsConfig := &dtls.Config{
		PSKIdentityHint: []byte(user),
		CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_GCM_SHA256},
		PSK: func(hint []byte) ([]byte, error) {
			return hex.DecodeString(secret)
		},
	}

	l._dtlsConn, err = dtls.DialWithContext(ctx, "udp", dtlsAddress, dtlsConfig)
	return err
}

// write appends the supplied bytes to HueDTSHeader and writes to the DTLS connection
func (l *Lightshow) write(b []byte) error {
	errFmt := "error writing to dtls connection: %w"

	if l._dtlsConn == nil {
		return fmt.Errorf(errFmt, ErrDTLSConnNil)
	}

	_, err := l._dtlsConn.Write(append(HueDTLSHeader, b...))

	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	return nil
}
