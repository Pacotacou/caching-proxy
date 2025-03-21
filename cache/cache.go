package cache

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"
)

// Cached HTTP response
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Timestamp  time.Time
}

// thread-safe in-memory cache for HTTP responses
type Cache struct {
	items map[string]CachedResponse
	mu    sync.RWMutex
}

// Create new Cache instance
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CachedResponse),
	}
}

// retrieve a cached response for the given key
func (c *Cache) Get(key string) (CachedResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, found := c.items[key]
	return resp, found
}

// store a response in the cache with the given key
func (c *Cache) Set(key string, statusCode int,
	headers http.Header, body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CachedResponse{
		StatusCode: statusCode,
		Headers:    headers.Clone(),
		Body:       body,
		Timestamp:  time.Now(),
	}

}

// remove all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]CachedResponse)
}

// returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// generate a unique cache key for an HTTP request
func CreateCacheKey(req *http.Request) string {
	return req.Method + ":" + req.URL.String()
}

// Write a cached response to an http.ResponseWriter
func (cr *CachedResponse) WriteToResponse(w http.ResponseWriter) {
	// Copy all headers from the cached response
	for key, values := range cr.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the X-Cache header to indicate a cache hit
	w.Header().Set("X-Cache", "HIT")

	// Set the Status Code
	w.WriteHeader(cr.StatusCode)

	// Write the body
	w.Write(cr.Body)
}

// copy an HTTP response for caching
func CopyResponse(resp *http.Response) (CachedResponse, error) {
	// Read and store the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CachedResponse{}, err
	}

	// Create a new reader from the body bytes for the original response
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// Create and return a CachedResponse
	return CachedResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
		Timestamp:  time.Now(),
	}, nil
}
