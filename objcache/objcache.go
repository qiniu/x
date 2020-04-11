package objcache

import (
	"sync"

	"github.com/qiniu/x/objcache/lru"
)

// Key type.
type Key = lru.Key

// Value type.
type Value = interface{}

// Context type.
type Context = interface{}

// OnEvictedFunc func.
type OnEvictedFunc = func(key Key, value Value)

// A GetterFunc implements Getter with a function.
type GetterFunc = func(ctx Context, key Key) (val Value, err error)

// newGroupHook, if non-nil, is called right after a new group is created.
var newGroupHook func(*Group)

// RegisterNewGroupHook registers a hook that is run each time
// a group is created.
func RegisterNewGroupHook(fn func(*Group)) {
	if newGroupHook != nil {
		panic("RegisterNewGroupHook called more than once")
	}
	newGroupHook = fn
}

// A Group is a cache namespace and associated data loaded spread over
// a group of 1 or more machines.
type Group struct {
	name string
	get  GetterFunc

	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// NewGroup creates a coordinated group-aware Getter from a Getter.
//
// The returned Getter tries (but does not guarantee) to run only one
// Get call at once for a given key across an entire set of peer
// processes. Concurrent callers both in the local process and in
// other processes receive copies of the answer once the original Get
// completes.
//
// The group name must be unique for each getter.
func NewGroup(name string, cacheNum int, getter GetterFunc, onEvicted ...OnEvictedFunc) *Group {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := groups[name]; dup {
		panic("duplicate registration of group " + name)
	}
	g := &Group{
		name: name,
		get:  getter,
	}
	g.mainCache.init(cacheNum, onEvicted...)
	if newGroupHook != nil {
		newGroupHook(g)
	}
	groups[name] = g
	return g
}

// Name returns the name of the group.
func (g *Group) Name() string {
	return g.name
}

// Get func.
func (g *Group) Get(ctx Context, key Key) (val Value, err error) {
	val, ok := g.mainCache.get(key)
	if ok {
		return
	}
	val, err = g.get(ctx, key)
	if err == nil {
		g.mainCache.add(key, val)
	}
	return
}

// TryGet func.
func (g *Group) TryGet(key Key) (val Value, ok bool) {
	return g.mainCache.get(key)
}

// CacheStats returns stats about the provided cache within the group.
func (g *Group) CacheStats() CacheStats {
	return g.mainCache.stats()
}

// cache is a wrapper around an *lru.Cache that adds synchronization,
// makes values always be ByteView, and counts the size of all keys and
// values.
type cache struct {
	mu         sync.RWMutex
	lru        *lru.Cache
	nhit, nget int64
}

func (c *cache) stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Items: c.itemsLocked(),
		Gets:  c.nget,
		Hits:  c.nhit,
	}
}

func (c *cache) init(cacheNum int, onEvicted ...OnEvictedFunc) {
	c.lru = lru.New(cacheNum)
	if onEvicted != nil {
		c.lru.OnEvicted = onEvicted[0]
	}
}

func (c *cache) add(key Key, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Add(key, value)
}

func (c *cache) get(key Key) (value Value, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nget++
	v, ok := c.lru.Get(key)
	if ok {
		value = v.(Value)
		c.nhit++
	}
	return
}

func (c *cache) items() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.itemsLocked()
}

func (c *cache) itemsLocked() int64 {
	return int64(c.lru.Len())
}

// CacheStats are returned by stats accessors on Group.
type CacheStats struct {
	Items int64
	Gets  int64
	Hits  int64
}
