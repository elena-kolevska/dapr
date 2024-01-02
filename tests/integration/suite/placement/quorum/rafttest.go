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

package quorum

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dapr/dapr/tests/integration/framework"
	"github.com/dapr/dapr/tests/integration/framework/process/daprd"
	prochttp "github.com/dapr/dapr/tests/integration/framework/process/http"
	"github.com/dapr/dapr/tests/integration/framework/process/placement"
	"github.com/dapr/dapr/tests/integration/framework/util"
	"github.com/dapr/dapr/tests/integration/suite"
	"github.com/stretchr/testify/require"
)

func init() {
	suite.Register(new(rafttest))
}

// rafttest tests placement can find quorum with tls disabled.
type rafttest struct {
	places []*placement.Placement
}

func (r *rafttest) Setup(t *testing.T) []framework.Option {
	// Create/Clear log files
	logs1, _ := os.OpenFile("/Users/elenakolevska/placement-work/logs1.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	logs2, _ := os.OpenFile("/Users/elenakolevska/placement-work/logs2.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	logs3, _ := os.OpenFile("/Users/elenakolevska/placement-work/logs3.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	memLogs, _ := os.OpenFile("/Users/elenakolevska/placement-work/go_memstats_alloc_bytes.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	logs1.Close()
	logs2.Close()
	logs3.Close()
	memLogs.Close()

	// Start three placement services
	fp := util.ReservePorts(t, 6)
	opts := []placement.Option{
		placement.WithInitialCluster(fmt.Sprintf("p1=localhost:%d,p2=localhost:%d,p3=localhost:%d", fp.Port(t, 0), fp.Port(t, 1), fp.Port(t, 2))),
		placement.WithInitialClusterPorts(fp.Port(t, 0), fp.Port(t, 1), fp.Port(t, 2)),
	}
	fmt.Printf("\n\n\nhttp://localhost:%s/placement/state\n\n", strconv.Itoa(fp.Port(t, 3)))
	fmt.Printf("\n\n\nhttp://localhost:%s/placement/state\n\n", strconv.Itoa(fp.Port(t, 4)))
	fmt.Printf("\n\n\nhttp://localhost:%s/placement/state\n\n", strconv.Itoa(fp.Port(t, 5)))

	r.places = []*placement.Placement{
		placement.New(t, append(opts, placement.WithID("p1"), placement.WithMetadataEnabled(true), placement.WithHealthzPort(fp.Port(t, 3)))...),
		placement.New(t, append(opts, placement.WithID("p2"), placement.WithMetadataEnabled(true), placement.WithHealthzPort(fp.Port(t, 4)))...),
		placement.New(t, append(opts, placement.WithID("p3"), placement.WithMetadataEnabled(true), placement.WithHealthzPort(fp.Port(t, 5)))...),
	}

	fp.Free(t)
	return []framework.Option{
		framework.WithProcesses(r.places[0], r.places[1], r.places[2]),
	}
}

func (r *rafttest) Run(t *testing.T, ctx context.Context) {
	r.places[0].WaitUntilRunning(t, ctx)
	r.places[1].WaitUntilRunning(t, ctx)
	r.places[2].WaitUntilRunning(t, ctx)

	client := util.HTTPClient(t)

	// Stream stats (memory usage or disk log size)
	go func() {
		for {
			//writeLogsMemory(client, r, t)
			writeLogsDisk()
			time.Sleep(time.Second)
		}

	}()

	//var types []string
	//numTypes := 2
	//for k := 0; k < numTypes; k++ {
	//	types = append(types, "myactortype-"+strconv.Itoa(k))
	//}

	// Bring sidecars up and down 100 times
	for j := 0; j < 20; j++ {
		var wg sync.WaitGroup
		numDaprSidecars := 9
		wg.Add(numDaprSidecars)
		// Start 9 Dapr sidecars with multiple actors each
		for i := 0; i < numDaprSidecars; i++ {

			go func(i int) {
				defer wg.Done()

				configFile := filepath.Join(t.TempDir(), "config.yaml")
				require.NoError(t, os.WriteFile(configFile, []byte(`
apiVersion: dapr.io/v1alpha1
kind: Configuration
metadata:
  name: actorstatettl
spec:
 features:
 - name: ActorStateTTL
   enabled: true
`), 0o600))
				var types []string

				// Simulate app exposing "dapr/config" endpoint with the list of actors
				handler := http.NewServeMux()
				handler.HandleFunc("/dapr/config", func(w http.ResponseWriter, r *http.Request) {
					types = []string{}
					numTypes := rand.Intn(200) + 1
					for k := 100 * i; k < 100*i+numTypes; k++ {
						types = append(types, "myactortype-"+strconv.Itoa(k))
					}

					w.Write([]byte(fmt.Sprintf(`{"entities": ["%s"]}`, strings.Join(types, `","`))))
				})
				handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(`OK`))
				})

				srv := prochttp.New(t, prochttp.WithHandler(handler))
				time.Sleep(2 * time.Second)

				// Starts a dapr sidecar
				daprd := daprd.New(t, daprd.WithResourceFiles(`
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: mystore
spec:
  type: state.in-memory
  version: v1
  metadata:
  - name: actorStateStore
    value: true
`),
					daprd.WithConfigs(configFile),
					daprd.WithPlacementAddresses(
						"localhost:"+strconv.Itoa(r.places[0].Port()),
						"localhost:"+strconv.Itoa(r.places[1].Port()),
						"localhost:"+strconv.Itoa(r.places[2].Port()),
					),
					daprd.WithAppPort(srv.Port()),
				)

				srv.Run(t, ctx)

				daprd.Run(t, ctx)
				daprd.WaitUntilRunning(t, ctx)
				daprd.WaitUntilAppHealth(t, ctx)
				fmt.Printf("\n\n\n-------------------Starting Dapr %d-------------------", i)

				// Check served actor types
				fmt.Println("\n\n\n-------------------Served actor types:-------------------")
				actorsUrl := "http://localhost:" + strconv.Itoa(srv.Port()) + "/dapr/config"
				fmt.Println(actorsUrl)
				reqA, err := http.NewRequest(http.MethodGet, actorsUrl, nil)
				require.NoError(t, err)

				respA, err := client.Do(reqA)
				require.NoError(t, err)

				bodyA, err := io.ReadAll(respA.Body)
				require.NoError(t, err)
				fmt.Println(string(bodyA))

				fmt.Println("\n\n\n-------------------Actors response-------------------")
				daprdURL := "http://localhost:" + strconv.Itoa(daprd.HTTPPort()) + "/v1.0/actors/" + types[0] + "/myactorid/method/foo"

				req, err := http.NewRequest(http.MethodGet, daprdURL, nil)
				require.NoError(t, err)

				resp, err := client.Do(req)
				require.NoError(t, err)

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				fmt.Println("\n\n\n-------------------Actors response-------------------")
				fmt.Println(string(body))
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				//assert.EventuallyWithT(t, func(c *assert.CollectT) {
				//	req, err := http.NewRequest(http.MethodGet, daprdURL, nil)
				//	require.NoError(t, err)
				//	resp, err := client.Do(req)
				//	require.NoError(t, err)
				//	body, err := io.ReadAll(resp.Body)
				//	require.NoError(t, err)
				//	fmt.Println("\n\n\n-------------------Actors response-------------------")
				//	fmt.Println(string(body))
				//	assert.Equal(c, http.StatusOK, resp.StatusCode)
				//	assert.NoError(t, resp.Body.Close())
				//}, time.Second*10, 100*time.Millisecond)

				// Sleep for random number between 0 and 5 seconds
				time.Sleep(time.Duration(rand.Intn(3)) * time.Second)

				daprd.Cleanup(t)
				srv.Cleanup(t)

				fmt.Printf("\n\n\n-------------------Killing Dapr %d-------------------", i)

			}(i)
		}
		wg.Wait()

		fmt.Printf("\n\n\n-------------------Cycle %d finished-------------------\n\n\n\n\n\n", j)
	}
}

func writeLogsMemory(client *http.Client, r *rafttest, t *testing.T) {
	resp, err := client.Get("http://localhost:" + strconv.Itoa(r.places[0].MetricsPort()))
	f, _ := os.OpenFile("/Users/elenakolevska/placement-work/go_memstats_alloc_bytes.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	if err == nil {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		for _, line := range strings.Split(string(body), "\n") {
			if strings.HasPrefix(line, "go_memstats_alloc_bytes") {
				valStr := strings.Replace(line, "go_memstats_alloc_bytes ", "", 1)
				bytes, _ := strconv.ParseFloat(valStr, 64)

				megabytes := bytes / 1e6 // Divide by 1e6 for conversion to megabytes

				fmt.Printf(">>%f\n", megabytes)
				f.WriteString(fmt.Sprintf("%f\n", megabytes))
				break
			}
		}
	}
}

func writeLogsDisk() {
	// Get file stats
	raft1, _ := os.Open("/Users/elenakolevska/placement-work/placementlogs-p1/wal-meta.db")
	raft2, _ := os.Open("/Users/elenakolevska/placement-work/placementlogs-p2/wal-meta.db")
	raft3, _ := os.Open("/Users/elenakolevska/placement-work/placementlogs-p3/wal-meta.db")

	logs1, _ := os.OpenFile("/Users/elenakolevska/placement-work/logs1.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	logs2, _ := os.OpenFile("/Users/elenakolevska/placement-work/logs2.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	logs3, _ := os.OpenFile("/Users/elenakolevska/placement-work/logs3.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	// Get file stats
	raft1Info, _ := raft1.Stat()
	raft2Info, _ := raft2.Stat()
	raft3Info, _ := raft3.Stat()

	// Get file size
	raft1fileSize := raft1Info.Size()
	raft2fileSize := raft2Info.Size()
	raft3fileSize := raft3Info.Size()

	sizeBytes1 := []byte(strconv.FormatInt(raft1fileSize, 10) + "\n")
	sizeBytes2 := []byte(strconv.FormatInt(raft2fileSize, 10) + "\n")
	sizeBytes3 := []byte(strconv.FormatInt(raft3fileSize, 10) + "\n")
	logs1.Write(sizeBytes1)
	logs2.Write(sizeBytes2)
	logs3.Write(sizeBytes3)

	raft1.Close()
	raft2.Close()
	raft3.Close()

	logs1.Close()
	logs2.Close()
	logs3.Close()
}
