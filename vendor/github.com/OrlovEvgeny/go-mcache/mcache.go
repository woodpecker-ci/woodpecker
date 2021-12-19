package mcache

import (
	"context"
	"github.com/OrlovEvgeny/go-mcache/gcmap"
	"github.com/OrlovEvgeny/go-mcache/item"
	"github.com/OrlovEvgeny/go-mcache/safeMap"
	"time"
)

const TTL_FOREVER = time.Hour * 87660

//initStore - returns context and context close func. Inited map storage and remove old cache
func (mc *CacheDriver) initStore() (context.Context, context.CancelFunc) {
	ctx, finish := context.WithCancel(context.Background())
	mc.storage = safeMap.NewStorage()
	mc.gc = gcmap.NewGC(ctx, mc.storage)
	return ctx, finish
}

//CacheDriver context struct
type CacheDriver struct {
	ctx      context.Context
	closeCtx context.CancelFunc
	storage  safeMap.SafeMap
	gc       *gcmap.GC
	instance *CacheDriver
}

//Deprecated: use New instead.
func StartInstance() *CacheDriver {
	cdriver := new(CacheDriver)
	ctx, finish := cdriver.initStore()
	cdriver.ctx = ctx
	cdriver.closeCtx = finish
	return cdriver
}

//New - returns CacheDriver struct
func New() *CacheDriver {
	cdriver := new(CacheDriver)
	ctx, finish := cdriver.initStore()
	cdriver.ctx = ctx
	cdriver.closeCtx = finish
	return cdriver
}

//Get - returns serialize data
func (mc *CacheDriver) Get(key string) (interface{}, bool) {
	data, ok := mc.storage.Find(key)
	if !ok {
		return item.Item{}.DataLink, false
	}
	entity := data.(item.Item)
	if entity.IsExpire() {
		return item.Item{}.DataLink, false
	}
	return entity.DataLink, true
}

//Set - add cache data value
func (mc *CacheDriver) Set(key string, value interface{}, ttl time.Duration) error {
	expire := time.Now().Local().Add(ttl)
	if ttl != TTL_FOREVER {
		go mc.gc.Expired(mc.ctx, key, ttl)
	}
	mc.storage.Insert(key, item.Item{Key: key, Expire: expire, DataLink: value})
	return nil
}

//Remove - value by key
func (mc *CacheDriver) Remove(key string) {
	mc.storage.Delete(key)
}

//Truncate - clean cache storage
func (mc *CacheDriver) Truncate() {
	mc.storage.Truncate()
}

//Len - returns current count storage
func (mc *CacheDriver) Len() int {
	return mc.storage.Len()
}

//GCBufferQueue - returns the current use len KeyChan chanel buffer
func (mc *CacheDriver) GCBufferQueue() int {
	return mc.gc.LenBufferKeyChan()
}

//Close - close all MCache
func (mc *CacheDriver) Close() map[string]interface{} {
	mc.closeCtx()
	return mc.storage.Close()
}

//Deprecated: use Set instead
func (mc *CacheDriver) SetPointer(key string, value interface{}, ttl time.Duration) error {
	return mc.Set(key, value, ttl)
}

//Deprecated: use Get instead
func (mc *CacheDriver) GetPointer(key string) (interface{}, bool) {
	return mc.Get(key)
}
