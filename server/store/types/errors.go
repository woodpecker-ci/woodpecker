// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"database/sql"
	"errors"
)

var (
	// RecordNotExist a Get or Update could not find the requested record.
	ErrRecordNotExist = sql.ErrNoRows

	// ErrInsertDuplicateDetected is returned when an insert fails because of unique constrains.
	ErrInsertDuplicateDetected = errors.New("on insert duplicate based on constraints was detected")

	// ErrInsertNone indicates that an insert did not create a record but statement itself was successful.
	ErrInsertNone = errors.New("no records where inserted")
)
