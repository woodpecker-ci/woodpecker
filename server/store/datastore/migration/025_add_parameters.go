package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var addParameters = xormigrate.Migration{
	ID: "add-parameters",
	MigrateSession: func(sess *xorm.Session) error {
		type parameters struct {
			ID           int64  `xorm:"pk autoincr 'parameter_id'"`
			RepoID       int64  `xorm:"UNIQUE(s) 'parameter_repo_id'"`
			Name         string `xorm:"UNIQUE(s) 'parameter_name'"`
			Branch       string `xorm:"UNIQUE(s) 'parameter_branch'"`
			Type         string `xorm:"'parameter_type'"`
			Description  string `xorm:"TEXT 'parameter_description'"`
			DefaultValue string `xorm:"TEXT 'parameter_default_value'"`
			TrimString   bool   `xorm:"'parameter_trim_string'"`
		}

		return sess.Sync(new(parameters))
	},
}
