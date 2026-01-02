package emotes

import (
	"container/list"
	"io"
	"net/http"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

const (
	maxCacheEntries = 500 // max emotes to keep in cache
)

type cacheEntry struct {
	key      string
	resource fyne.Resource
	loading  bool
	element  *list.Element // for LRU tracking
}

// thread safe caching for emote images with LRU eviction
type ImageCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	lru     *list.List // front = most recently used
	client  *http.Client
}

// create a new image cache
func NewImageCache() *ImageCache {
	return &ImageCache{
		entries: make(map[string]*cacheEntry),
		lru:     list.New(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// return a cache if available, updates LRU position
func (c *ImageCache) Get(key string) (fyne.Resource, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok || entry.resource == nil {
		return nil, false
	}
	// move to front (most recently used)
	if entry.element != nil {
		c.lru.MoveToFront(entry.element)
	}
	return entry.resource, true
}

// return cached resource or starts an async load
// If loading is started, onLoaded will be called when complete
// Returns (resource, isLoading). If isLoading is true, the resource is a placeholder
func (c *ImageCache) GetOrLoad(key, url string, onLoaded func(fyne.Resource)) (fyne.Resource, bool) {
	c.mu.Lock()

	entry, ok := c.entries[key]
	if ok {
		if entry.resource != nil {
			// move to front (most recently used)
			if entry.element != nil {
				c.lru.MoveToFront(entry.element)
			}
			c.mu.Unlock()
			return entry.resource, false
		}
		if entry.loading {
			c.mu.Unlock()
			return nil, true
		}
	}

	// start loading
	entry = &cacheEntry{key: key, loading: true}
	c.entries[key] = entry
	c.mu.Unlock()

	go c.loadAsync(key, url, onLoaded)

	return nil, true
}

func (c *ImageCache) loadAsync(key, url string, onLoaded func(fyne.Resource)) {
	resp, err := c.client.Get(url)
	if err != nil {
		c.markFailed(key)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.markFailed(key)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.markFailed(key)
		return
	}

	resource := fyne.NewStaticResource(key, data)

	c.mu.Lock()
	if entry, ok := c.entries[key]; ok {
		entry.resource = resource
		entry.loading = false
		// add to LRU list
		entry.element = c.lru.PushFront(entry)
		// evict oldest entries if cache is full
		c.evictOldest()
	}
	c.mu.Unlock()

	if onLoaded != nil {
		onLoaded(resource)
	}
}

// evictOldest removes least recently used entries until cache is within limit
// must be called with lock held
func (c *ImageCache) evictOldest() {
	for c.lru.Len() > maxCacheEntries {
		oldest := c.lru.Back()
		if oldest == nil {
			break
		}
		entry := oldest.Value.(*cacheEntry)
		c.lru.Remove(oldest)
		delete(c.entries, entry.key)
	}
}

func (c *ImageCache) markFailed(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok {
		entry.loading = false
	}
}

// clear removes all cached entries
func (c *ImageCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
	c.lru = list.New()
}
