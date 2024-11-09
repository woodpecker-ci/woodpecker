// Copyright 2023 Woodpecker Authors
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

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

func (s storage) ServerConfigGet(key string) (string, error) {
	config := new(model.ServerConfig)
	err := wrapGet(s.engine.ID(key).Get(config))
	if err != nil {
		return "", err
	}

	return config.Value, nil
}

func (s storage) ServerConfigSet(key, value string) error {
	config := &model.ServerConfig{
		Key: key,
	}

	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	count, err := sess.Count(config)
	if err != nil {
		return err
	}

	config.Value = value

	if count == 0 {
		_, err = sess.Insert(config)
	} else {
		_, err = sess.Where("`key` = ?", config.Key).Cols("value").Update(config)
	}
	if err != nil {
		return err
	}

	return sess.Commit()
}

func (s storage) ServerConfigDelete(key string) error {
	config := &model.ServerConfig{
		Key: key,
	}

	return wrapDelete(s.engine.Delete(config))
}
