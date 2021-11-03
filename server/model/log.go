package model

type Logs struct {
	ID     int64  `xorm:"pk autoincr 'log_id'"`
	ProcID int64  `xorm:"log_job_id"`
	Data   []byte `xorm:"log_data"`
}
