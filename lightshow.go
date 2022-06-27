package lightshow

import (
	"context"
	"fmt"
	"time"

	"github.com/pion/dtls/v2"
)

type Lightshow struct {
	_dtlsConn *dtls.Conn
}

func (l *Lightshow) RunMap(ctx context.Context, m Map) error {
	ctx, done := context.WithCancel(ctx)
	defer done()

	m.SortEvents()

	startTime := time.Now()

	for _, event := range m.Events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After((time.Second * time.Duration(event.StartTime))):
			go event.run(ctx, l)

			go func() {
				fmt.Printf("triggered event after %f\r\n", time.Since(startTime).Seconds())
			}()
		}
	}

	return nil
}

func (l *Lightshow) RunMapAsync(ctx context.Context, m Map) <-chan error {
	c := make(chan error, 1)

	go func() {
		c <- l.RunMap(ctx, m)
	}()

	return c
}
