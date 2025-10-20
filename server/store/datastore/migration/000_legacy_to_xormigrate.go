package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var legacyToXormigrate = xormigrate.Migration{
	ID: "legacy-to-xormigrate",
	MigrateSession: func(sess *xorm.Session) error {
		type migrations struct {
			Name string `xorm:"UNIQUE"`
		}

		var mig []*migrations
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
