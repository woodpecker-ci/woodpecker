// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"errors"
	"maps"

	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func (s storage) agentSnapshotFindByHash(sess *xorm.Session, hash string) (*model.AgentSnapshot, error) {
	snapshot := new(model.AgentSnapshot)
	if err := wrapGet(sess.Where("hash = ?", hash).Get(snapshot)); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// AgentSnapshotPersist stores the snapshot unless an identical one exists, and returns
// the stored one.
func (s storage) AgentSnapshotPersist(snapshot *model.AgentSnapshot) (*model.AgentSnapshot, error) {
	hash, err := snapshot.ContentHash()
	if err != nil {
		return nil, err
	}

	// no transaction: it would not stop a concurrent insert anyway, and on postgres a
	// failed insert aborts the transaction, which would break the lookup below
	sess := s.engine.NewSession()
	defer sess.Close()

	existing, err := s.agentSnapshotFindByHash(sess, hash)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, types.ErrRecordNotExist) {
		return nil, err
	}

	// do not change the caller's snapshot
	stored := *snapshot
	stored.ID = 0
	stored.Hash = hash
	stored.AgentID = 0
	stored.CustomLabels = maps.Clone(snapshot.CustomLabels)
	err = wrapInsert(sess.Insert(&stored))
	if errors.Is(err, types.ErrInsertDuplicateDetected) {
		// a concurrent request persisted the identical snapshot after our lookup, so use
		// that one (same race as https://github.com/woodpecker-ci/woodpecker/issues/7173)
		return s.agentSnapshotFindByHash(sess, hash)
	}
	if err != nil {
		return nil, err
	}

	return &stored, nil
}

func (s storage) AgentSnapshotFind(id int64) (*model.AgentSnapshot, error) {
	snapshot := new(model.AgentSnapshot)
	return snapshot, wrapGet(s.engine.ID(id).Get(snapshot))
}
