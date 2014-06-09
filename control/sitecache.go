package control

import (
	"database/sql"
	"fmt"
	"github.com/mrmorphic/goss/cache"
	"github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"net/http"
	"time"
)

// siteCacheEntry is the primary structure in the site cache, which works by periodically
// fetching all site tree records that are used in routing and navigation, and caching them
// in a data structure that makes for fast retrieval.
type siteCacheEntry struct {
	ID           int
	ClassName    string
	ParentID     int
	Title        string
	MenuTitle    string
	URLSegment   string
	RelativePath string
}

type SiteCache struct {
	// raw list of site tree records
	raw []*siteCacheEntry

	// a map of relative site paths to siteCacheEntry objects
	paths map[string]*siteCacheEntry

	// a map of object IDs to data objects.
	objByID map[int]interface{}
}

// primeSiteCache is responsible for re-computing the data structures in the cache. It does this
// by requerying the database, rebuilding the structure, and finally replacing the data structures
// atomically. In this way, a request being processed will either get the old version or the new
// version, but whichever version it's using won't be replaced mid-request.
func primeSiteCache() (*SiteCache, error) {
	r, e := orm.Query(`select "ID","ClassName","ParentID","Title","MenuTitle","URLSegment" from "SiteTree_Live"`)
	defer r.Close()

	if e != nil {
		fmt.Printf("ERROR EXECUTING SQL: %s\n", e)
		return nil, e
	}

	newCache := newSiteCache()

	for r.Next() {
		e := newCache.ReadRow(r)
		if e != nil {
			return nil, e
		}
	}

	newCache.derivePaths()

	fmt.Printf("primeSiteCache: %s\n", newCache)

	return newCache, nil
}

func newSiteCache() *SiteCache {
	return &SiteCache{raw: []*siteCacheEntry{}, paths: map[string]*siteCacheEntry{}, objByID: map[int]interface{}{}}
}

func getSiteCache() *SiteCache {
	key := "goss.Sitetree"
	result := cache.Get(key)
	if result != nil {
		return result.(*SiteCache)
	}

	c, _ := primeSiteCache()

	if configuration.cacheSiteTreeNavTTL > 0 {
		cache.Store(key, c, time.Duration(configuration.cacheSiteTreeNavTTL)*time.Second)
	}

	return c
}

// After computing the cache, if content controller subsequently loads a page for rendering against,
// it can add this to the cache. It will be cleared when the site tree cache is next cleared.
func (c *SiteCache) CacheDataObject(id int, object interface{}) {
	c.objByID[id] = object
}

func (c *SiteCache) GetCacheByID(id int) interface{} {
	return c.objByID[id]
}

func (c *SiteCache) ReadRow(r *sql.Rows) error {
	cols, e := r.Columns()
	colCount := len(cols)

	var field []interface{}
	for i := 0; i < colCount; i++ {
		switch {
		case cols[i][:2] == "b:":
			field = append(field, new(sql.NullBool))
		case cols[i][:2] == "f:":
			field = append(field, new(sql.NullFloat64))
		case cols[i][:2] == "i:":
			field = append(field, new(sql.NullInt64))
		case cols[i][:2] == "s:":
			field = append(field, new(sql.NullString))
		case cols[i][:2] == "t:":
			field = append(field, new(time.Time))
		default:
			field = append(field, new(sql.NullString))
		}
	}

	e = r.Scan(field...)

	if e != nil {
		fmt.Printf("got an error though: %s\n", e)
		return e
	}

	m := &siteCacheEntry{}

	for i, c := range cols {
		v := flatten(field[i])
		data.Set(m, c, v)
	}

	c.raw = append(c.raw, m)

	return nil
}

// Given a set of siteCacheEntry objects in c.raw, derive the map of paths
func (c *SiteCache) derivePaths() {
	for _, entry := range c.raw {
		c.derivePathEntry(entry)
	}
}

func (c *SiteCache) derivePathEntry(entry *siteCacheEntry) {
	key := ""
	if entry.ParentID == 0 {
		// if not parent, key is just URLSegment, home normalised to /
		key = entry.URLSegment
		if key == "home" {
			key = "/"
		}
	} else {
		// locate the parent
		// key is parent key + this URL Segment
		parent := c.findRawByID(entry.ParentID)
		if parent == nil {
			// shouldn't happen, but may if referential integrity is broken (parentID refers to non-existent parent); treat
			// it like ParentID==0
			key = entry.URLSegment
		} else {
			// ensure parent key is defined
			if parent.RelativePath == "" {
				c.derivePathEntry(parent)
			}
			key = parent.RelativePath + "/" + entry.URLSegment
		}
	}
	entry.RelativePath = key
	c.paths[key] = entry
}

func (c *SiteCache) findRawByID(id int) *siteCacheEntry {
	for _, entry := range c.raw {
		if entry.ID == id {
			return entry
		}
	}
	return nil
}

func flatten(sqlField interface{}) interface{} {
	switch v := sqlField.(type) {
	case *sql.NullBool:
		return v.Bool
	case *sql.NullFloat64:
		return v.Float64
	case *sql.NullInt64:
		return v.Int64
	case *sql.NullString:
		return v.String
	}

	return sqlField
}

// Given a request, find the site tree entry by path and return the ID
func (c *SiteCache) findPageToRender(r *http.Request) (int, bool) {
	p := c.paths[r.URL.Path]
	if p != nil {
		return p.ID, true
	}
	return 0, false
}
