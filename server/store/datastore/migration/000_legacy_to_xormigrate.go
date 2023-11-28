package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type v000Migrations struct {
	Name string `xorm:"UNIQUE"`
}

func (m *v000Migrations) TableName() string {
	return "migrations"
}

var legacyToXormigrate = xormigrate.Migration{
	ID: "legacy-to-xormigrate",
	MigrateSession: func(sess *xorm.Session) error {
		var mig []*v000Migrations
		if err := sess.Find(&mig); err != nil {
			return err
		}
		for _, m := range mig {
			if _, err := sess.Insert(&xormigrate.Migration{
				ID: m.Name,
			}); err != nil {
				return err
			}
		}

		return sess.DropTable("migrations")
	},
}
