package cache

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type ExpiredHandler func(dm map[string]interface{})

type Item struct {
	Object     interface{}
	Expiration *time.Time
}

// Returns true if the item has expired.
func (item *Item) Expired() bool {
	if item.Expiration == nil {
		return false
	}
	return item.Expiration.Before(time.Now())
}

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

type Cache struct {
	*cache
	// If this is confusing, see the comment at the bottom of New()
}

type cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	items             map[string]*Item
	janitor           *janitor
	expiredHandler    ExpiredHandler
}

func (c *cache) RegExpiredHandler(eh ExpiredHandler) {
	c.expiredHandler = eh
}

// Add an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *cache) Set(k string, x interface{}, d time.Duration) {
	c.Lock()
	c.set(k, x, d)
	// TODO: Calls to mu.Unlock are currently not deferred because defer
	// adds ~200 ns (as of go1.)
	c.Unlock()
}

func (c *cache) set(k string, x interface{}, d time.Duration) {
	var e *time.Time
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		t := time.Now().Add(d)
		e = &t
	}
	c.items[k] = &Item{
		Object:     x,
		Expiration: e,
	}
}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *cache) Add(k string, x interface{}, d time.Duration) error {
	c.Lock()
	_, found := c.get(k)
	if found {
		c.Unlock()
		return fmt.Errorf("Item %s already exists", k)
	}
	c.set(k, x, d)
	c.Unlock()

	return nil
}

// Refresh expiration attribute for the cache key only if it already exists.
// Returns an error otherwise.
func (c *cache) Refresh(k string, d time.Duration) bool {
	c.Lock()
	defer c.Unlock()

	item, found := c.items[k]
	if !found || item.Expired() {
		return false
	}
	var e *time.Time
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		t := time.Now().Add(d)
		e = &t
	}
	item.Expiration = e
	return true
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *cache) Get(k string) (interface{}, bool) {
	c.RLock()
	x, found := c.get(k)
	c.RUnlock()
	return x, found
}

func (c *cache) get(k string) (interface{}, bool) {
	item, found := c.items[k]
	if !found || item.Expired() {
		return nil, false
	}
	return item.Object, true
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (c *cache) Delete(k string) {
	c.Lock()
	c.delete(k)
	c.Unlock()
}

func (c *cache) delete(k string) {
	delete(c.items, k)
}

// Delete all expired items from the cache.
func (c *cache) DeleteExpired() {
	c.Lock()
	var hasExpiredItem bool
	var dm map[string]interface{}
	for k, v := range c.items {
		if v.Expired() {
			hasExpiredItem = true
			if dm == nil {
				dm = make(map[string]interface{})
			}
			dm[k] = v.Object
			c.delete(k)
		}
	}
	c.Unlock()

	if hasExpiredItem && c.expiredHandler != nil {
		go c.expiredHandler(dm)
	}
}

// Write the cache's items (using Gob) to an io.Writer.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	c.RLock()
	defer c.RUnlock()
	for _, v := range c.items {
		gob.Register(v.Object)
	}
	err = enc.Encode(&c.items)
	return
}

// Save the cache's items to the given filename, creating the file if it
// doesn't exist, and overwriting it if it does.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) SaveFile(fname string) error {
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = c.Save(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Add (Gob-serialized) cache items from an io.Reader, excluding any items with
// keys that already exist (and haven't expired) in the current cache.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]*Item{}
	err := dec.Decode(&items)
	if err == nil {
		c.Lock()
		defer c.Unlock()
		for k, v := range items {
			ov, found := c.items[k]
			if !found || ov.Expired() {
				c.items[k] = v
			}
		}
	}
	return err
}

// Load and add cache items from the given filename, excluding any items with
// keys that already exist in the current cache.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = c.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Returns the items in the cache. This may include items that have expired,
// but have not yet been cleaned up. If this is significant, the Expiration
// fields of the items should be checked. Note that explicit synchronization
// is needed to use a cache and its corresponding Items() return value at
// the same time, as the map is shared.
func (c *cache) Items() map[string]*Item {
	c.RLock()
	defer c.RUnlock()
	return c.items
}

// Returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up. Equivalent to len(c.Items()).
func (c *cache) ItemCount() int {
	c.RLock()
	n := len(c.items)
	c.RUnlock()
	return n
}

// Delete all items from the cache.
func (c *cache) Flush() {
	c.Lock()
	c.items = map[string]*Item{}
	c.Unlock()
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *cache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			c.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
	}
	c.janitor = j
	go j.Run(c)
}

func newCache(de time.Duration, m map[string]*Item) *cache {
	if de == 0 {
		de = -1
	}
	c := &cache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newCacheWithJanitor(de time.Duration, ci time.Duration, m map[string]*Item) *Cache {
	c := newCache(de, m)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	C := &Cache{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

// Return a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling c.DeleteExpired().
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]*Item)
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}
