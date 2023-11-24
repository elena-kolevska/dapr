package errutil

import (
	"fmt"
	kitErrors "github.com/dapr/kit/errors"
	grpcCodes "google.golang.org/grpc/codes"
	"net/http"
)

func NewErrStateStoreNotConfigured() *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.FailedPrecondition,
		http.StatusInternalServerError,
		"state store is not configured",
		"ERR_STATE_STORE_NOT_CONFIGURED",
	)
}

func NewErrStateStoreNotFound(storeName string) *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.InvalidArgument, // TODO check if it was used in the past, we should change it. It should be grpcCodes.NotFound
		http.StatusBadRequest,
		fmt.Sprintf("state store %s is not found", storeName),
		"ERR_STATE_STORE_NOT_FOUND",
	).
		WithErrorInfo(StateStore+ErrNotFound, nil)
}

func NewErrStateStoreQueryFailed(storeName string, detail string) *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.Internal,
		http.StatusInternalServerError,
		fmt.Sprintf("state store %s query failed: %s", storeName, detail),
		"ERR_STATE_QUERY",
	)
}

func NewErrStateStoreQueryUnsupported() *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.Internal,
		http.StatusInternalServerError,
		"state store does not support querying",
		"ERR_STATE_STORE_NOT_SUPPORTED",
	)
}

func NewErrStateStoreTooManyTransactionalOps(count int, max int) *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.InvalidArgument,
		http.StatusBadRequest,
		fmt.Sprintf("the transaction contains %d operations, which is more than what the state store supports: %d", count, max),
		"ERR_STATE_STORE_TOO_MANY_TRANSACTIONS",
	)
}

func NewErrStateStoreInvalidKeyName(key string, daprSeparator string) *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.InvalidArgument,
		http.StatusBadRequest,
		fmt.Sprintf("input key/keyPrefix '%s' can't contain '%s'", key, daprSeparator),
		"",
	).WithErrorInfo(StateStore+IllegalKey, nil)
}
