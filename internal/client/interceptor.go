package client

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/kassie.v1.SessionService/Login":       true,
	"/kassie.v1.SessionService/Refresh":     true,
	"/kassie.v1.SessionService/GetProfiles": true,
}

func (c *Client) authInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if publicMethods[method] {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		ctx = c.attachToken(ctx)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}

		if !isAuthError(err) {
			return err
		}

		if refreshErr := c.Refresh(ctx); refreshErr != nil {
			return err
		}

		ctx = c.attachToken(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *Client) attachToken(ctx context.Context) context.Context {
	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	if token == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func (c *Client) needsRefresh() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.accessToken == "" {
		return false
	}

	return time.Until(c.expiresAt) < 60*time.Second
}

func isAuthError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	if st.Code() == codes.Unauthenticated {
		return true
	}

	msg := strings.ToLower(st.Message())
	return strings.Contains(msg, "token") && strings.Contains(msg, "expired")
}
