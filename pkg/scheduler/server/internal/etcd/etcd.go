/*
Copyright 2025 The Dapr Authors
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

package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"

	"github.com/dapr/dapr/pkg/healthz"
	"github.com/dapr/dapr/pkg/modes"
	"github.com/dapr/dapr/pkg/security"
	"github.com/dapr/kit/logger"
)

var log = logger.NewLogger("dapr.scheduler.server.etcd")

const NODE0 = "dapr-scheduler-server-0"
const NODE1 = "dapr-scheduler-server-1"
const NODE2 = "dapr-scheduler-server-2"

type Options struct {
	Name                string
	InitialCluster      []string
	ClientPort          uint64
	SpaceQuota          int64
	CompactionMode      string
	CompactionRetention string
	SnapshotCount       uint64
	MaxSnapshots        uint
	MaxWALs             uint
	Security            security.Handler

	DataDir string
	Healthz healthz.Healthz
	Mode    modes.DaprMode
}

type Interface interface {
	Run(ctx context.Context) error
	Client(ctx context.Context) (*clientv3.Client, error)
}

type etcd struct {
	mode modes.DaprMode

	etcd   *embed.Etcd
	client *clientv3.Client
	config *embed.Config

	readyCh chan struct{}
	hz      healthz.Target
}

func New(opts Options) (Interface, error) {
	config, err := config(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd config: %w", err)
	}

	return &etcd{
		hz:      opts.Healthz.AddTarget(),
		config:  config,
		readyCh: make(chan struct{}),
		mode:    opts.Mode,
	}, nil
}

func (e *etcd) Run(ctx context.Context) error {
	defer e.hz.NotReady()
	log.Info("Starting etcd")

	var err error

	initialPeers := make(map[string]string, 3)
	split := strings.Split(e.config.InitialCluster, ",")
	if len(split) != 3 {
		return fmt.Errorf("invalid initial cluster: %s", e.config.InitialCluster)
	}

	for _, s := range split {
		nsplit := strings.Split(s, "=")
		if len(nsplit) != 2 {
			return fmt.Errorf("invalid initial cluster: %s", e.config.InitialCluster)
		}
		initialPeers[nsplit[0]] = nsplit[1]
	}

	//if e.mode == modes.KubernetesMode {
	if true {
		if e.config.Name == NODE0 {
			e.config.ClusterState = embed.ClusterStateFlagNew
			e.etcd, err = e.handleKube0(ctx, initialPeers)
		} else if e.config.Name == NODE1 {
			fmt.Printf("Starting etcd on %s\n", NODE1)

			e.config.ClusterState = embed.ClusterStateFlagExisting
			e.config.InitialCluster = fmt.Sprintf("%s=%s,%s=%s", NODE0, initialPeers[NODE0], NODE1, initialPeers[NODE1])

			e.etcd, err = e.startEtcdWithTimeout()
		} else if e.config.Name == NODE2 {
			fmt.Printf("Starting etcd on %s\n", NODE2)

			e.config.ClusterState = embed.ClusterStateFlagExisting
			e.config.InitialCluster = fmt.Sprintf("%s=%s,%s=%s,%s=%s", NODE0, initialPeers[NODE0], NODE1, initialPeers[NODE1], NODE2, initialPeers[NODE2])

			e.etcd, err = e.startEtcdWithTimeout()
		} else {
			return fmt.Errorf("invalid node name: %s", e.config.Name)
		}
		if err != nil {
			return fmt.Errorf("error starting etcd: %w", err)
		}

	} else {
		e.etcd, err = embed.StartEtcd(e.config)
		if err != nil {
			return fmt.Errorf("failed to start etcd: %w", err)
		}
	}

	//e.client, err = clientv3.New(clientv3.Config{
	//	Endpoints: []string{e.config.ListenClientUrls[0].Host},
	//	Logger:    e.etcd.GetLogger(),
	//})
	//if err != nil {
	//	return errors.Join(err, e.client.Close())
	//}
	e.client, err = e.getClientWithTimeout()
	if err != nil {
		return err
	}

	select {
	case <-e.etcd.Server.ReadyNotify():
		log.Info("Etcd server is ready!")
	case <-ctx.Done():
		return ctx.Err()
	}

	e.hz.Ready()
	close(e.readyCh)

	select {
	case err := <-e.etcd.Err():
		return err
	case <-ctx.Done():
		return nil
	}
}

func (e *etcd) startEtcdWithTimeout() (*embed.Etcd, error) {
	t := time.NewTimer(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			return nil, fmt.Errorf("********* failed to start etcd within 30 seconds")
		default:
			fmt.Println("********* Starting etcd with", e.config.InitialCluster)
			etcd, err := embed.StartEtcd(e.config)
			if err == nil {
				fmt.Println("********* Etcd started successfully!", e.config.InitialCluster)
				return etcd, nil
			}

			fmt.Printf("********* Failed to start etcd: %v, retrying...\n", err)

			time.Sleep(1 * time.Second)
		}
	}
}

func (e *etcd) getClientWithTimeout() (*clientv3.Client, error) {
	t := time.NewTimer(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			return nil, fmt.Errorf("********* failed to get etcd client within 30 seconds")
		default:
			client, err := clientv3.New(clientv3.Config{
				Endpoints: []string{e.config.ListenClientUrls[0].Host},
				Logger:    e.etcd.GetLogger(),
			})
			if err == nil {
				fmt.Printf("***** Client connected to %s\n", e.config.ListenClientUrls[0].Host)
				return client, nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func (e *etcd) handleKube0(ctx context.Context, initialPeers map[string]string) (*embed.Etcd, error) {
	existingCluster := filepath.Join(e.config.Dir, "dapr-existing-cluster")
	_, err := os.Stat(existingCluster)
	if err == nil {
		log.Infof("Found existing cluster at %s", existingCluster)
		return nil, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	e.config.InitialCluster = fmt.Sprintf("%s=%s", NODE0, initialPeers[NODE0])

	etcd, err := embed.StartEtcd(e.config)
	if err != nil {
		return nil, err
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{e.config.ListenClientUrls[0].Host},
		Logger:    etcd.GetLogger(),
	})
	if err != nil {
		return nil, errors.Join(err, e.client.Close())
	}

	//members, err := client.MemberList(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if len(members.Members) > 1 {
	//	fmt.Println("******* Cluster already exists, skipping member add")
	//	return nil, os.WriteFile(existingCluster, nil, 0o600)
	//}

	// Add node 2
	if _, err := client.MemberAdd(ctx, []string{initialPeers[NODE1]}); err != nil {
		// TODO check if error is about node already being a member and return if it is
		return nil, err
	}

	fmt.Println("******* Added second node to cluster")

	// Wait until Node 2 joins and is ready
	if err := backoff.Retry(func() error {
		members, err := client.MemberList(ctx)
		if err != nil {
			return err
		}

		vd, _ := json.MarshalIndent(members, "", "  ")
		fmt.Println("******* cluster members: ")
		fmt.Println(string(vd))

		if len(members.Members) != 2 {
			fmt.Printf("******* Cluster not ready yet, retrying... (%d members)", len(members.Members))
			return fmt.Errorf("cluster not ready yet")
		}

		resp, err := client.Status(ctx, members.Members[1].ClientURLs[0])
		fmt.Printf("*********************** NODE %s STATUS ***********************", members.Members[1].GetName())

		vd, _ = json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(vd))

		if err != nil {
			return fmt.Errorf("node %s is unhealthy: %v", members.Members[1].GetName(), err)
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx)); err != nil {
		log.Errorf("Failed adding second member: %s", err)
		return nil, err
	}

	// Add node 3
	if _, err := client.MemberAdd(ctx, []string{initialPeers[NODE2]}); err != nil {
		return nil, err
	}
	fmt.Println("******* Added node 3 to cluster")

	return etcd, os.WriteFile(existingCluster, nil, 0o600)
}

func (e *etcd) Client(ctx context.Context) (*clientv3.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-e.readyCh:
		return e.client, nil
	}
}

func (e *etcd) Close() error {
	defer log.Info("Etcd shut down")

	var err error
	if e.client != nil {
		//ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		//defer cancel()
		//if _, err = e.client.MemberRemove(ctx, uint64(e.etcd.Server.ID())); err != nil {
		//	log.Errorf("Failed to remove member from cluster during shutdown: %s", err)
		//}
		err = e.client.Close()
	}

	if e.etcd != nil {
		e.etcd.Close()
	}

	return err
}
