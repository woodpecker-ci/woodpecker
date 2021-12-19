package gcmap

import (
	"context"
	"github.com/OrlovEvgeny/go-mcache/safeMap"
	"sync"
	"time"
)

var (
	gcInstance   *GC
	loadInstance = false
)

//keyset - sync slice for expired keys
type keyset struct {
	keys  [2][]string
	cur   int
	mutex sync.Mutex
}

func (kset *keyset) len() int {
	return len(kset.keys[kset.cur])
}

func (kset *keyset) append(key string) {
	kset.mutex.Lock()
	defer kset.mutex.Unlock()

	kset.keys[kset.cur] = append(kset.keys[kset.cur], key)
}

func (kset *keyset) swap() []string {
	kset.mutex.Lock()
	defer kset.mutex.Unlock()

	keys := kset.keys[kset.cur]
	kset.keys[kset.cur] = kset.keys[kset.cur][:0]

	kset.cur = (kset.cur + 1) & 0x1

	return keys
}

//GC garbage clean struct
type GC struct {
	storage safeMap.SafeMap
	keyChan chan string
}

//NewGC - singleton func, returns *GC struct
func NewGC(ctx context.Context, store safeMap.SafeMap) *GC {
	if loadInstance {
		return gcInstance
	}

	gc := new(GC)
	gc.storage = store
	gc.keyChan = make(chan string, 10000)
	go gc.ExpireKey(ctx)
	gcInstance = gc
	loadInstance = true

	return gc
}

//LenBufferKeyChan - returns len usage buffet of keyChan chanel
func (gc GC) LenBufferKeyChan() int {
	return len(gc.keyChan)
}

//ExpireKey - collects foul keys, what to remove later
func (gc GC) ExpireKey(ctx context.Context) {
	kset := &keyset{cur: 0}
	kset.keys[0] = make([]string, 0, 100)
	kset.keys[1] = make([]string, 0, 100)

	go gc.heartBeatGC(ctx, kset)

	for {
		select {
		case key := <-gc.keyChan:
			kset.append(key)

		case <-ctx.Done():
			loadInstance = false
			return
		}
	}
}

//heartBeatGC removes old keys by timer
func (gc GC) heartBeatGC(ctx context.Context, kset *keyset) {
	//TODO it may be worthwhile to set a custom interval for deleting old keys
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ticker.C:
			if kset.len() == 0 {
				continue
			}
			keys := kset.swap()
			gc.storage.Flush(keys)

		case <-ctx.Done():
			return
		}
	}
}

//Expired - fund Expired, gorutine which is launched every time the method is called, and ensures that the key is removed from the repository after the time expires
func (gc GC) Expired(ctx context.Context, key string, duration time.Duration) {
	select {
	case <-time.After(duration):
		gc.keyChan <- key
		return
	case <-ctx.Done():
		return
	}
}
