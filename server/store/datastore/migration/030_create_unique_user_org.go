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

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// checks whether the user_org_id (table users) in table orgs for a given user_id is unique in table orgs
// if it is not unique, a new entry in table orgs is being created and user_org_id in table users is updated accordingly for user_id
// relates to a migration issue on Codeberg: https://codeberg.org/Codeberg-CI/feedback/issues/149#issuecomment-1546709
var createUniqueUserOrg = xormigrate.Migration{
	ID: "createUniqueUserOrg",
	Migrate: func(tx *xorm.Engine) error {
		_, err := tx.Exec(`
		DO
		$do$
		DECLARE
		    _user_id int;
		    _user_org_id int;
		BEGIN
		    FOR _user_id, _user_org_id IN (SELECT user_id, user_org_id FROM users)
		    LOOP
			IF (SELECT COUNT(*) FROM orgs WHERE user_id = _user_id AND user_org_id = _user_org_id) > 1 THEN
			    INSERT INTO orgs (user_id, user_org_id) VALUES (_user_id, _user_org_id);
			    UPDATE users SET user_org_id = (SELECT MAX(user_org_id) FROM orgs WHERE user_id = _user_id) WHERE user_id = _user_id;
			END IF;
		    END LOOP;
		END
		$do$
	    `)
		return err
	},
}
