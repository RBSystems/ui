package bff

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/lazarette/lazarette"
	"github.com/byuoitav/ui/log"
	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var (
	grpcInitCreds sync.Once
	grpcCreds     credentials.TransportCredentials
)

// LazaretteState represents the current lazarette state in a room
type LazaretteState struct {
	*sync.Map
}

type lazMessage struct {
	Key  string
	Data interface{}
}

func createGrpcConn(ctx context.Context, addr string, ssl bool) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}

	if ssl {
		grpcInitCreds.Do(setupGrpcCreds)
		opts = append(opts, grpc.WithTransportCredentials(grpcCreds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create grpc connection: %s", err)
	}

	// TODO reconnect

	return conn, nil
}

func (c *Client) updateLazaretteState(laz lazarette.LazaretteClient) {
	for {
		select {
		case <-c.kill:
			return
		case message := <-c.lazUpdates:
			c.stats.Lazarette.UpdatesSent++

			data, err := json.Marshal(message.Data)
			if err != nil {
				c.Warn("unable to marshal lazarette message", zap.String("key", message.Key), zap.Error(err))
				continue
			}

			c.Debug("Storing key in lazarette", zap.String("key", message.Key), zap.ByteString("data", data))

			// store it in our local map
			c.lazs.Store(message.Key, data)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			// cancel needs to be called before this block ends!!!

			_, err = laz.Set(ctx, &lazarette.KeyValue{
				Timestamp: ptypes.TimestampNow(),
				Key:       c.roomID + message.Key,
				Data:      data,
			})
			if err != nil {
				c.Warn("unable to set updated key to lazarette", zap.String("key", message.Key), zap.Error(err))
				cancel()
			}

			cancel()
		}
	}
}

func (c *Client) subLazaretteState(sub lazarette.Lazarette_SubscribeClient) {
	for {
		select {
		case <-c.kill:
			return
		default:
			kv, err := sub.Recv()
			switch {
			case errors.Is(err, io.EOF):
				c.Warn("lazarette stream ended", zap.Error(err))
				return
			case err != nil:
				s := status.Convert(err)
				switch s.Code() {
				case codes.Canceled, codes.DeadlineExceeded:
					c.Warn("ending lazarette stream", zap.Error(s.Err()))
					return
				default:
					c.Warn("lazarette stream error", zap.Error(err))
					continue
				}
			case kv == nil:
				continue
			}

			c.stats.Lazarette.UpdatesRecieved++

			// strip off beginning roomID so that we only have the actual key
			key := strings.TrimPrefix(kv.GetKey(), c.roomID)

			updateRoom := false

			// stick the value into our map
			switch {
			case strings.HasPrefix(key, lazSharing):
				c.Debug("Got lazarette update", zap.String("key", key), zap.ByteString("data", kv.GetData()))

				var data lazShareData
				if err := json.Unmarshal(kv.GetData(), &data); err != nil {
					c.Warn("unable to parse share data from lazarette", zap.String("key", key), zap.Error(err))
					continue
				}
				c.lazs.Store(key, data)
				updateRoom = true
			default:
			}

			if updateRoom {
				roomMsg, err := JSONMessage("room", c.GetRoom())
				if err != nil {
					c.Warn("unable to build updated room: %w", zap.Error(err))
					continue
				}

				c.Out <- roomMsg
			}
		}
	}
}

func setupGrpcCreds() {
	pool, err := x509.SystemCertPool()
	if err != nil {
		// should only warn on windows? i think?
		log.P.Fatal("unable to get system cert pool", zap.Error(err))
	}

	grpcCreds = credentials.NewClientTLSFromCert(pool, "")
}
