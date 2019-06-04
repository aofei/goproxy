package goproxy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"time"
)

// ErrCacheNotFound is the error resulting if a path search failed to find a
// cache.
var ErrCacheNotFound = errors.New("cache not found")

// Cacher is the interface that defines a set of methods used to cache module
// files for the `Goproxy`.
//
// Note that the cache names must be UNIX-style paths.
//
// If you are looking for some useful implementations of the `Cacher`, simply
// visit the "github.com/goproxy/goproxy/cachers" package.
type Cacher interface {
	// Cache returns the matched `Cache` for the name from the underlying
	// cacher. It returns the `ErrCacheNotFound` if not found.
	Cache(ctx context.Context, name string) (Cache, error)

	// SetCache sets the c to the underlying cacher.
	SetCache(ctx context.Context, c Cache) error
}

// Cache is the cache unit of the `Cacher`.
type Cache interface {
	io.Reader
	io.Seeker
	io.Closer

	// Name returns the unique name of the underlying cache.
	//
	// Note that the returned name must be a UNIX-style path.
	Name() string

	// Size returns the length in bytes of the underlying cache.
	Size() int64

	// ModTime returns the modification time of the underlying cache.
	ModTime() time.Time

	// Checksum returns the xxHash (XXH64) checksum of the underlying cache.
	Checksum() []byte
}

// tempCacher implements the `Cacher` by using the `map[string]tempCache`.
type tempCacher struct {
	caches map[string]*tempCache
}

// Cache implements the `Cacher`.
func (tc *tempCacher) Cache(ctx context.Context, name string) (Cache, error) {
	c, ok := tc.caches[name]
	if !ok {
		return nil, ErrCacheNotFound
	}

	return c, nil
}

// SetCache implements the `Cacher`.
func (tc *tempCacher) SetCache(ctx context.Context, c Cache) error {
	b, err := ioutil.ReadAll(c)
	if err != nil {
		return err
	}

	tc.caches[c.Name()] = &tempCache{
		readSeeker: bytes.NewReader(b),
		name:       c.Name(),
		size:       c.Size(),
		modTime:    c.ModTime(),
		checksum:   c.Checksum(),
	}

	return nil
}

// tempCache implements the `Cache`. It is the cache unit of the `tempCacher`.
type tempCache struct {
	readSeeker io.ReadSeeker
	name       string
	size       int64
	modTime    time.Time
	checksum   []byte
}

// Read implements the `Cache`.
func (tc *tempCache) Read(b []byte) (int, error) {
	return tc.readSeeker.Read(b)
}

// Seek implements the `Cache`.
func (tc *tempCache) Seek(offset int64, whence int) (int64, error) {
	return tc.readSeeker.Seek(offset, whence)
}

// Close implements the `Cache`.
func (tc *tempCache) Close() error {
	_, err := tc.readSeeker.Seek(0, io.SeekEnd)
	return err
}

// Name implements the `Cache`.
func (tc *tempCache) Name() string {
	return tc.name
}

// Size implements the `Cache`.
func (tc *tempCache) Size() int64 {
	return tc.size
}

// ModTime implements the `Cache`.
func (tc *tempCache) ModTime() time.Time {
	return tc.modTime
}

// Checksum implements the `Cache`.
func (tc *tempCache) Checksum() []byte {
	return tc.checksum
}
