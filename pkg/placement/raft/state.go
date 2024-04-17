/*
Copyright 2021 The Dapr Authors
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

package raft

import (
	"fmt"
	"io"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-msgpack/v2/codec"

	"github.com/dapr/dapr/pkg/placement/hashing"
)

var ErrNamespaceNotFound = fmt.Errorf("namespace not found")

// DaprHostMember represents Dapr runtime actor host member which serve actor types.
type DaprHostMember struct {
	// Name is the unique name of Dapr runtime host.
	Name string
	// AppID is Dapr runtime app ID.
	AppID string

	// Namespace is the namespace of the Dapr runtime host.
	Namespace string

	// Entities is the list of Actor Types which this Dapr runtime supports.
	Entities []string

	// UpdatedAt is the last time when this host member info is updated.
	UpdatedAt int64

	// Version of the Actor APIs supported by the Dapr runtime
	APILevel uint32
}

type DaprNamespace struct {
	// Members includes Dapr runtime hosts.
	Members map[string]*DaprHostMember

	// hashingTableMap is the map for storing consistent hashing data
	// per Actor types. This will be generated when log entries are replayed.
	// While snapshotting the state, this member will not be saved. Instead,
	// hashingTableMap will be recovered in snapshot recovery process.
	hashingTableMap map[string]*hashing.Consistent
}

type DaprHostMemberStateData struct {
	// Index is the index number of raft log.
	Index uint64
	// Version of the actor APIs for the cluster
	APILevel uint32
	// TableGeneration is the generation of hashingTableMap.
	// This is increased whenever hashingTableMap is updated.
	TableGeneration uint64
	Namespace       map[string]*DaprNamespace
}

func newDaprHostMemberStateData() DaprHostMemberStateData {
	return DaprHostMemberStateData{
		Namespace: make(map[string]*DaprNamespace),
	}
}

// DaprHostMemberState is the state to store Dapr runtime host and
// consistent hashing tables.
type DaprHostMemberState struct {
	lock sync.RWMutex

	config DaprHostMemberStateConfig

	data DaprHostMemberStateData
}

type DaprHostMemberStateConfig struct {
	replicationFactor int64
	minAPILevel       uint32
	maxAPILevel       uint32
}

func newDaprHostMemberState(config DaprHostMemberStateConfig) *DaprHostMemberState {
	return &DaprHostMemberState{
		config: config,
		data:   newDaprHostMemberStateData(),
	}
}

func (s *DaprHostMemberState) Index() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data.Index
}

// APILevel returns the current API level of the cluster.
func (s *DaprHostMemberState) APILevel() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data.APILevel
}

func (s *DaprHostMemberState) Members(ns string) (map[string]*DaprHostMember, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	n, ok := s.data.Namespace[ns]
	if !ok {
		return nil, ErrNamespaceNotFound
	}
	members := make(map[string]*DaprHostMember, len(n.Members))
	for k, v := range n.Members {
		members[k] = v
	}
	return members, nil
}

func (s *DaprHostMemberState) TableGeneration() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data.TableGeneration
}

// Internal function that updates the API level in the object.
// The API level can only be increased.
// Make sure you have a lock before calling this method.
func (s *DaprHostMemberState) updateAPILevel() {
	var observedMinLevel uint32

	// Loop through all namespaces and members to find the minimum API level
	for _, ns := range s.data.Namespace {
		for k := range ns.Members {
			apiLevel := ns.Members[k].APILevel
			if apiLevel <= 0 {
				apiLevel = 0
			}
			if observedMinLevel == 0 || observedMinLevel > apiLevel {
				observedMinLevel = apiLevel
			}
		}
	}

	// Only enforce minAPILevel if value > 0
	// 0 is the default value of the struct.
	// -1 is the default value of the CLI flag.
	if s.config.minAPILevel >= uint32(0) && observedMinLevel < s.config.minAPILevel {
		observedMinLevel = s.config.minAPILevel
	}

	// Only enforce maxAPILevel if value > 0
	// 0 is the default value of the struct.
	// -1 is the default value of the CLI flag.
	if s.config.maxAPILevel >= uint32(0) && observedMinLevel > s.config.maxAPILevel {
		observedMinLevel = s.config.maxAPILevel
	}

	if observedMinLevel > s.data.APILevel {
		s.data.APILevel = observedMinLevel
	}
}

func (s *DaprHostMemberState) hashingTableMap(ns string) (map[string]*hashing.Consistent, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	n, ok := s.data.Namespace[ns]
	if !ok {
		return nil, ErrNamespaceNotFound
	}

	return n.hashingTableMap, nil
}

func (s *DaprHostMemberState) clone() *DaprHostMemberState {
	s.lock.RLock()
	defer s.lock.RUnlock()

	newMembers := &DaprHostMemberState{
		config: s.config,
		data: DaprHostMemberStateData{
			Index:           s.data.Index,
			Namespace:       make(map[string]*DaprNamespace, len(s.data.Namespace)),
			TableGeneration: s.data.TableGeneration,
			APILevel:        s.data.APILevel,
		},
	}

	for nsName, nsData := range s.data.Namespace {
		newMembers.data.Namespace[nsName] = &DaprNamespace{
			Members:         make(map[string]*DaprHostMember, len(nsData.Members)),
			hashingTableMap: make(map[string]*hashing.Consistent, len(nsData.hashingTableMap)),
		}
		for k, v := range nsData.Members {
			m := &DaprHostMember{
				Name:      v.Name,
				AppID:     v.AppID,
				Entities:  make([]string, len(v.Entities)),
				UpdatedAt: v.UpdatedAt,
				APILevel:  v.APILevel,
			}
			copy(m.Entities, v.Entities)
			newMembers.data.Namespace[nsName].Members[k] = m
		}
	}

	return newMembers
}

// caller should hold lock.
func (s *DaprHostMemberState) updateHashingTables(host *DaprHostMember) {
	// TODO: @elena - come back to this and check if we can rely on s.data.Namespace[host.Namespace] existing
	for _, e := range host.Entities {
		if _, ok := s.data.Namespace[host.Namespace].hashingTableMap[e]; !ok {
			s.data.Namespace[host.Namespace].hashingTableMap[e] = hashing.NewConsistentHash(s.config.replicationFactor)
		}

		s.data.Namespace[host.Namespace].hashingTableMap[e].Add(host.Name, host.AppID, 0)
	}
}

// caller should holds lock.
func (s *DaprHostMemberState) removeHashingTables(host *DaprHostMember) {
	for _, e := range host.Entities {
		if t, ok := s.data.Namespace[host.Namespace].hashingTableMap[e]; ok {
			t.Remove(host.Name)

			// if no dedicated actor service instance for the particular actor type,
			// we must delete consistent hashing table to avoid the memory leak.
			if len(t.Hosts()) == 0 {
				delete(s.data.Namespace[host.Namespace].hashingTableMap, e)
			}
		}
	}
}

// upsertMember upserts member host info to the FSM state and returns true
// if the hashing table update happens.
func (s *DaprHostMemberState) upsertMember(host *DaprHostMember) bool {
	if !s.isActorHost(host) {
		return false
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	ns, ok := s.data.Namespace[host.Namespace]
	if !ok {
		s.data.Namespace[host.Namespace] = &DaprNamespace{
			Members: make(map[string]*DaprHostMember),
		}
		ns = s.data.Namespace[host.Namespace]
	}

	if m, ok := ns.Members[host.Name]; ok {
		// No need to update consistent hashing table if the same dapr host member exists
		if m.AppID == host.AppID && m.Name == host.Name && cmp.Equal(m.Entities, host.Entities) {
			m.UpdatedAt = host.UpdatedAt
			return false
		}

		// Remove hashing table because the existing member is invalid
		// and needs to be updated by new member info.
		s.removeHashingTables(m)
	}

	ns.Members[host.Name] = &DaprHostMember{
		Name:      host.Name,
		AppID:     host.AppID,
		UpdatedAt: host.UpdatedAt,
		APILevel:  host.APILevel,
	}

	// Update hashing table only when host reports actor types
	ns.Members[host.Name].Entities = make([]string, len(host.Entities))
	copy(ns.Members[host.Name].Entities, host.Entities)

	s.updateHashingTables(ns.Members[host.Name])
	s.updateAPILevel()

	// Increase hashing table generation version. Runtime will compare the table generation
	// version with its own and then update it if it is new.
	s.data.TableGeneration++

	return true
}

// removeMember removes members from membership and update hashing table and returns true
// if hashing table update happens.
func (s *DaprHostMemberState) removeMember(host *DaprHostMember) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	ns, ok := s.data.Namespace[host.Namespace]
	if !ok {
		return false
	}

	if m, ok := ns.Members[host.Name]; ok {
		s.removeHashingTables(m)
		s.data.TableGeneration++
		delete(ns.Members, host.Name)
		s.updateAPILevel()

		return true
	}

	return false
}

func (s *DaprHostMemberState) isActorHost(host *DaprHostMember) bool {
	return len(host.Entities) > 0
}

// caller should hold lock.
func (s *DaprHostMemberState) restoreHashingTables() {
	for _, ns := range s.data.Namespace {
		if ns.hashingTableMap == nil {
			ns.hashingTableMap = map[string]*hashing.Consistent{}
		}

		for _, m := range ns.Members {
			s.updateHashingTables(m)
		}
	}
}

func (s *DaprHostMemberState) restore(r io.Reader) error {
	dec := codec.NewDecoder(r, &codec.MsgpackHandle{})
	var data DaprHostMemberStateData
	if err := dec.Decode(&data); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data

	s.restoreHashingTables()
	s.updateAPILevel()
	return nil
}

func (s *DaprHostMemberState) persist(w io.Writer) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	b, err := marshalMsgPack(s.data)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}
