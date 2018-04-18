package cache

import (
	"testing"

	"github.com/turbinelabs/test/assert"
)

func TestNewLRU(t *testing.T) {
	c, err := NewLRU(0)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "positive size")

	c, err = NewLRU(10)
	assert.Nil(t, err)
	assert.NonNil(t, c)
	assert.NonNil(t, c.(*lruCache).lru)
}

func TestLRUCacheBasicOperations(t *testing.T) {
	c, err := NewLRU(10)
	assert.Nil(t, err)

	assert.False(t, c.Add("k1", "v1"))
	assert.False(t, c.Add("k2", "v2"))
	assert.False(t, c.Add("k3", "v3"))
	assert.True(t, c.Add("k1", "v1-again"))
	assert.Equal(t, c.Len(), 3)

	assert.True(t, c.Remove("k3"))
	assert.Equal(t, c.Len(), 2)

	assert.False(t, c.Remove("never-added"))
	assert.Equal(t, c.Len(), 2)

	v1, ok1 := c.Get("k1")
	assert.True(t, ok1)
	assert.Equal(t, v1, "v1-again")

	v2, ok2 := c.Get("k2")
	assert.True(t, ok2)
	assert.Equal(t, v2, "v2")

	c.Clear()
	assert.Equal(t, c.Len(), 0)
}

func TestLRUCacheForEach(t *testing.T) {
	c, err := NewLRU(5)
	assert.Nil(t, err)

	c.Add("k1", "v1")
	c.Add("k2", "v2")
	c.Add("k3", "v3")
	c.Add("k4", "v4")
	c.Add("k5", "v5")

	kvs := [][]string{}
	c.ForEach(func(k, v interface{}) {
		kvs = append(kvs, []string{k.(string), v.(string)})
	})
	assert.ArrayEqual(t, kvs, [][]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
		{"k4", "v4"},
		{"k5", "v5"},
	})

	c.Get("k1")

	kvs = [][]string{}
	c.ForEach(func(k, v interface{}) {
		kvs = append(kvs, []string{k.(string), v.(string)})
	})
	assert.ArrayEqual(t, kvs, [][]string{
		{"k2", "v2"},
		{"k3", "v3"},
		{"k4", "v4"},
		{"k5", "v5"},
		{"k1", "v1"},
	})
}
