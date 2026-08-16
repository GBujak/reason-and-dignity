package db

import (
	"fmt"

	"github.com/maypok86/otter"
	"golang.org/x/sync/singleflight"
)

type SqliteConnectionCache struct {
	cache       otter.Cache[string, *SqliteConnection]
	singleGroup singleflight.Group
}

func NewSqliteConnectionCache(capacity int) (*SqliteConnectionCache, error) {
	cache, err := otter.MustBuilder[string, *SqliteConnection](capacity).
		DeletionListener(func(key string, value *SqliteConnection, cause otter.DeletionCause) {
			go func() {
				_ = value.Close()
			}()
		}).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to initialize otter cache: %w", err)
	}

	return &SqliteConnectionCache{
		cache: cache,
	}, nil
}

func (c *SqliteConnectionCache) Get(path string) (*SqliteConnection, error) {
	if conn, ok := c.cache.Get(path); ok {
		return conn, nil
	}

	res, err, _ := c.singleGroup.Do(path, func() (interface{}, error) {
		if conn, ok := c.cache.Get(path); ok {
			return conn, nil
		}

		conn, err := createClient(path)
		if err != nil {
			return nil, err
		}

		c.cache.Set(path, conn)
		return conn, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*SqliteConnection), nil
}
