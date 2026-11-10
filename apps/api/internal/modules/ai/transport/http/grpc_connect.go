package http

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

func connectAgent(ctx context.Context, target string, options []grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(target, options...)
	if err != nil {
		return nil, err
	}
	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return conn, nil
		}
		if state == connectivity.Idle {
			conn.Connect()
		}
		if !conn.WaitForStateChange(ctx, state) {
			conn.Close()
			return nil, ctx.Err()
		}
	}
}
