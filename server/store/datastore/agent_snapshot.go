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

// WorkflowUpdateWithSnapshot persists the snapshot and its workflow reference together.
func (s storage) WorkflowUpdateWithSnapshot(workflow *model.Workflow, snapshot *model.AgentSnapshot) error {
	hash, err := snapshot.ContentHash()
	if err != nil {
		return err
	}

	return retryWrite(func() error {
		sess := s.engine.NewSession()
		defer sess.Close()
		if err := sess.Begin(); err != nil {
			return err
		}

		// Lock an existing snapshot before reading it. A no-op write also acquires
		// SQLite's write lock, and waits for concurrent cleanup to commit first.
		if err := wrapUpdate(sess.Where("hash = ?", hash).SetExpr("hash", "hash").Update(new(model.AgentSnapshot))); err != nil {
			return err
		}
		stored, err := s.agentSnapshotFindByHash(sess, hash)
		if errors.Is(err, types.ErrRecordNotExist) {
			// Do not change the caller's snapshot.
			insert := *snapshot
			insert.ID = 0
			insert.Hash = hash
			insert.AgentID = 0
			insert.CustomLabels = maps.Clone(snapshot.CustomLabels)
			stored = &insert
			err = wrapInsert(sess.Insert(stored))
		}
		if err != nil {
			return err
		}

		workflow.AgentSnapshotID = stored.ID
		if err := wrapUpdate(sess.ID(workflow.ID).AllCols().Update(workflow)); err != nil {
			return err
		}
		return wrapUpdate(0, sess.Commit())
	})
}

func (s storage) AgentSnapshotFind(id int64) (*model.AgentSnapshot, error) {
	snapshot := new(model.AgentSnapshot)
	return snapshot, wrapGet(s.engine.ID(id).Get(snapshot))
}
