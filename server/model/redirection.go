package model

type Redirection struct {
	ID       int64  `xorm:"pk autoincr 'redirection_id'"`
	RepoID   int64  `xorm:"'repo_id'"`
	FullName string `xorm:"UNIQUE INDEX 'repo_full_name'"`
}
