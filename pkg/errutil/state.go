/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	).WithErrorInfo(StateStore+ErrNotConfigured, nil)
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

func NewErrStateStoreQueryUnsupported() *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.Internal,
		http.StatusInternalServerError,
		"state store does not support querying",
		"ERR_STATE_STORE_NOT_SUPPORTED",
	).WithErrorInfo(StateStore+"QUERYING_"+ErrNotSupported, nil)
}
func NewErrStateStoreQueryFailed(storeName string, detail string) *kitErrors.Error {
	return kitErrors.New(
		grpcCodes.Internal,
		http.StatusInternalServerError,
		fmt.Sprintf("state store %s query failed: %s", storeName, detail),
		"ERR_STATE_QUERY",
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
