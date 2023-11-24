/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implieh.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/status"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	commonv1 "github.com/dapr/dapr/pkg/proto/common/v1"
	rtv1 "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/dapr/tests/integration/framework"
	procdaprd "github.com/dapr/dapr/tests/integration/framework/process/daprd"
	"github.com/dapr/dapr/tests/integration/suite"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpcCodes "google.golang.org/grpc/codes"
)

var i = 0

func init() {
	suite.Register(new(stateErrors))
}

type stateErrors struct {
	daprd *procdaprd.Daprd
}

func (b *stateErrors) Setup(t *testing.T) []framework.Option {
	b.daprd = procdaprd.New(t, procdaprd.WithResourceFiles(`
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: mystore
spec:
  type: state.in-memory
  version: v1
`))
	i = i + 1

	return []framework.Option{
		framework.WithProcesses(b.daprd),
	}
}

func (b *stateErrors) Run(t *testing.T, ctx context.Context) {
	b.daprd.WaitUntilRunning(t, ctx)

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", b.daprd.GRPCPort()), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	client := rtv1.NewDaprClient(conn)

	// Covers errutils.NewErrStateStoreNotFound()
	t.Run("state store doesn't exist", func(t *testing.T) {
		req := &rtv1.SaveStateRequest{
			StoreName: "mystore-doesnt-exist",
			States:    []*commonv1.StateItem{{Value: []byte("value1")}},
		}
		_, err = client.SaveState(ctx, req)
		require.Error(t, err)

		s, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, grpcCodes.InvalidArgument, s.Code())
		require.Equal(t, fmt.Sprintf("state store %s is not found", "mystore-doesnt-exist"), s.Message())

		//Check status details
		require.Equal(t, 1, len(s.Details()))
		errInfo := s.Details()[0]
		require.IsType(t, &errdetails.ErrorInfo{}, errInfo)
		require.Equal(t, "DAPR_STATE_NOT_FOUND", errInfo.(*errdetails.ErrorInfo).GetReason())
		require.Equal(t, framework.Domain, errInfo.(*errdetails.ErrorInfo).GetDomain())
		require.Nil(t, errInfo.(*errdetails.ErrorInfo).GetMetadata())
	})

	// Covers errutils.NewErrStateStoreInvalidKeyName()
	t.Run("invalid key name", func(t *testing.T) {
		keyName := "invalid||key"

		req := &rtv1.SaveStateRequest{
			StoreName: "mystore",
			States:    []*commonv1.StateItem{{Key: keyName, Value: []byte("value1")}},
		}
		_, err = client.SaveState(ctx, req)
		require.Error(t, err)

		s, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, grpcCodes.InvalidArgument, s.Code())
		require.Equal(t, fmt.Sprintf("input key/keyPrefix '%s' can't contain '||'", keyName), s.Message())

		//Check status details
		require.Equal(t, 1, len(s.Details()))
		errInfo := s.Details()[0]
		require.IsType(t, &errdetails.ErrorInfo{}, errInfo)
		require.Equal(t, "DAPR_STATE_ILLEGAL_KEY", errInfo.(*errdetails.ErrorInfo).GetReason())
		require.Equal(t, framework.Domain, errInfo.(*errdetails.ErrorInfo).GetDomain())
		require.Nil(t, errInfo.(*errdetails.ErrorInfo).GetMetadata())
	})

}
