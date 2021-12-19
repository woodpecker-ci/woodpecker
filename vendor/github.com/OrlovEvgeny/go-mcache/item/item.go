package item

import "time"

//Item - data storage structure
type Item struct {
	Key      string
	Expire   time.Time
	Data     []byte
	DataLink interface{}
}

//IsExpire check expire cache, return true if the time of cache is expired
func (i Item) IsExpire() bool {
	return i.Expire.Before(time.Now().Local())
}
