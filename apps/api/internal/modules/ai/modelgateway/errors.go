package modelgateway

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayError struct {
	GRPCCode   codes.Code
	PublicCode string
	Err        error
}

func (e GatewayError) Error() string {
	if e.Err == nil {
		return e.PublicCode
	}
	return e.PublicCode + ": " + e.Err.Error()
}

func (e GatewayError) Unwrap() error {
	return e.Err
}

func typedProviderError(publicCode string, err error) error {
	return GatewayError{GRPCCode: codes.Unavailable, PublicCode: publicCode, Err: err}
}

func schemaError(err error) error {
	return GatewayError{GRPCCode: codes.FailedPrecondition, PublicCode: "model_schema_invalid", Err: err}
}

func grpcError(err error) error {
	var gatewayErr GatewayError
	if errors.As(err, &gatewayErr) {
		return status.Error(gatewayErr.GRPCCode, gatewayErr.PublicCode)
	}
	return status.Error(codes.Internal, "model_gateway_error")
}

func invalidArgument(format string, args ...any) error {
	return status.Error(codes.InvalidArgument, fmt.Sprintf(format, args...))
}
