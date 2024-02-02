package migration

import (
	"os"
	"testing"

	"xorm.io/xorm"
)

type User struct {
	UserID    int `xorm:"'user_id'"`
	UserOrgID int `xorm:"'user_org_id'"`
}

func TestCreateUniqueUserOrg(t *testing.T) {
	// Set up the test database
	config := os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
	db, err := xorm.NewEngine("postgres", config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, _ = db.Exec(`
            CREATE TABLE users (
                    user_id INTEGER PRIMARY KEY,
                    user_org_id INTEGER
            );
            CREATE TABLE orgs (
                id INTEGER PRIMARY KEY,
                is_user CHARACTER
        );
    `)

	// Insert mock non-unique user_org_id entries
	_, _ = db.Exec(`
            INSERT INTO users (user_id, user_org_id) VALUES
            (1, 1),
            (2, 2),
            (2, 1);
    `)
	_, _ = db.Exec(`
            INSERT INTO orgs (id, is_user) VALUES
            (1, 1),
            (2, 1);
    `)

	// Run the migration
	err = createUniqueUserOrg.Migrate(db)
	if err != nil {
		t.Fatalf("Failed to run migration: %v", err)
	}

	// Check that the database is in the expected state
	var count int
	err = db.DB().QueryRow("SELECT COUNT(*) FROM users GROUP BY user_id, user_org_id HAVING COUNT(*) > 1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count > 0 {
		t.Errorf("Found %d users with more than one org, want 0", count)
	}
}
