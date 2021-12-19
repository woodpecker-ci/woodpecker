package safeMap

import "github.com/OrlovEvgeny/go-mcache/item"

//safeMap -  data command chanel
type safeMap chan commandData

// commandAction - action constant
type commandAction int

//commandData - data for safe communication with the storage
type commandData struct {
	action commandAction
	key    string
	keys   []string
	value  interface{}
	result chan<- interface{}
	data   chan<- map[string]interface{}
}

//Auto inc command
const (
	REMOVE commandAction = iota
	FLUSH
	FIND
	INSERT
	COUNT
	TRUNCATE
	END
)

//findResult - result data, returns to Find() method
type findResult struct {
	value interface{}
	found bool
}

//SafeMap the main interface for communicating with the storage
type SafeMap interface {
	Insert(string, interface{})
	Delete(string)
	Truncate()
	Flush([]string)
	Find(string) (interface{}, bool)
	Len() int
	Close() map[string]interface{}
}

//NewStorage - make new storage, return interface SafeMap
func NewStorage() SafeMap {
	sm := make(safeMap)
	go sm.run()
	return sm
}

//run - gorutina with a safe map
func (sm safeMap) run() {
	store := make(map[string]interface{})
	for command := range sm {
		switch command.action {
		case INSERT:
			store[command.key] = command.value
		case REMOVE:
			delete(store, command.key)
		case FLUSH:
			flush(store, command.keys)
		case FIND:
			value, found := store[command.key]
			command.result <- findResult{value, found}
		case COUNT:
			command.result <- len(store)
		case TRUNCATE:
			clearMap(store)
		case END:
			close(sm)
			command.data <- store
		}
	}
}

//Insert - add value to storage
func (sm safeMap) Insert(key string, value interface{}) {
	sm <- commandData{action: INSERT, key: key, value: value}
}

//Delete - delete value from storage
func (sm safeMap) Delete(key string) {
	sm <- commandData{action: REMOVE, key: key}
}

//Flush - delete many keys from storage
func (sm safeMap) Flush(keys []string) {
	sm <- commandData{action: FLUSH, keys: keys}
}

//Find - find storage, returns findResult struct
func (sm safeMap) Find(key string) (value interface{}, found bool) {
	reply := make(chan interface{})
	sm <- commandData{action: FIND, key: key, result: reply}
	result := (<-reply).(findResult)
	return result.value, result.found
}

//Len - returns current count storage value
func (sm safeMap) Len() int {
	reply := make(chan interface{})
	sm <- commandData{action: COUNT, result: reply}
	return (<-reply).(int)
}

//Close - close storage and return storage map
func (sm safeMap) Close() map[string]interface{} {
	reply := make(chan map[string]interface{})
	sm <- commandData{action: END, data: reply}
	return <-reply
}

//Truncate - clean storage
func (sm safeMap) Truncate() {
	sm <- commandData{action: TRUNCATE}
}

//clearMap - helper by Truncate func
func clearMap(store map[string]interface{}) {
	for k := range store {
		delete(store, k)
	}
}

//flush - helper by Flush func
func flush(s map[string]interface{}, keys []string) {
	for _, v := range keys {
		value, ok := s[v]
		if !ok {
			continue
		}
		if value.(item.Item).IsExpire() {
			delete(s, v)
		}
	}
}
