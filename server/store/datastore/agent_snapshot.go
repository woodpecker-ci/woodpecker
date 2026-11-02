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

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

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

		// Lock an existing snapshot before reading it, so concurrent cleanup commits first.
		if err := agentSnapshotsLock(sess.Where("hash = ?", hash)); err != nil {
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

// agentSnapshotsDeleteUnused deletes the snapshots referenced only by workflows of
// this pipeline. It must run in the transaction that deletes these workflows.
func (s storage) agentSnapshotsDeleteUnused(sess *xorm.Session, pipelineID int64) error {
	var ids []int64
	// Ascending IDs keep the lock order equal for concurrent cleanups.
	if err := sess.Table(new(model.Workflow)).Distinct("agent_snapshot_id").
		Where("pipeline_id = ? AND agent_snapshot_id > 0", pipelineID).
		OrderBy("agent_snapshot_id").Find(&ids); err != nil {
		return err
	}

	for _, id := range ids {
		// Wait for an assignment that is about to reference this snapshot.
		if err := agentSnapshotsLock(sess.ID(id)); err != nil {
			return err
		}
		usedElsewhere := sess.Select("1").Where(builder.Eq{"agent_snapshot_id": id}.
			And(builder.Neq{"pipeline_id": pipelineID}))
		if s.engine.Dialect().URI().DBType != schemas.SQLITE {
			// Use a current read even under MySQL's default repeatable-read isolation.
			usedElsewhere = usedElsewhere.ForUpdate()
		}
		used, err := usedElsewhere.Exist(new(model.Workflow))
		if err != nil {
			return wrapUpdate(0, err)
		}
		if used {
			continue
		}
		if _, err := sess.ID(id).Delete(new(model.AgentSnapshot)); err != nil {
			return err
		}
	}
	return nil
}

// agentSnapshotsLock locks the snapshots matched by sess with a no-op write,
// which also acquires SQLite's write lock.
func agentSnapshotsLock(sess *xorm.Session) error {
	return wrapUpdate(sess.SetExpr("hash", "hash").Update(new(model.AgentSnapshot)))
}

func (s storage) AgentSnapshotFind(id int64) (*model.AgentSnapshot, error) {
	snapshot := new(model.AgentSnapshot)
	return snapshot, wrapGet(s.engine.ID(id).Get(snapshot))
}
