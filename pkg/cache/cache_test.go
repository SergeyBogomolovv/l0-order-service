package cache

import (
	"context"
	"testing"
	"time"
)

func TestLRUCache(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		ttl      time.Duration
		actions  func(c *LRUCache, t *testing.T)
	}{
		{
			name:     "set and get within TTL",
			capacity: 2,
			ttl:      time.Second,
			actions: func(c *LRUCache, t *testing.T) {
				c.Set("a", []byte("1"))
				if v, ok := c.Get("a"); !ok || string(v) != "1" {
					t.Errorf("expected value=1, got=%v, ok=%v", v, ok)
				}
			},
		},
		{
			name:     "get after expiration",
			capacity: 2,
			ttl:      time.Millisecond * 50,
			actions: func(c *LRUCache, t *testing.T) {
				c.Set("a", []byte("1"))
				time.Sleep(time.Millisecond * 60)
				if _, ok := c.Get("a"); ok {
					t.Errorf("expected key to be expired")
				}
			},
		},
		{
			name:     "evict oldest when over capacity",
			capacity: 2,
			ttl:      time.Second,
			actions: func(c *LRUCache, t *testing.T) {
				c.Set("a", []byte("1"))
				c.Set("b", []byte("2"))
				c.Set("c", []byte("3"))
				if _, ok := c.Get("a"); ok {
					t.Errorf("expected key 'a' to be evicted")
				}
				if v, ok := c.Get("b"); !ok || string(v) != "2" {
					t.Errorf("expected b=2, got %v", v)
				}
				if v, ok := c.Get("c"); !ok || string(v) != "3" {
					t.Errorf("expected c=3, got %v", v)
				}
			},
		},
		{
			name:     "update value resets TTL",
			capacity: 2,
			ttl:      time.Millisecond * 50,
			actions: func(c *LRUCache, t *testing.T) {
				c.Set("a", []byte("1"))
				time.Sleep(time.Millisecond * 30)
				c.Set("a", []byte("2"))
				time.Sleep(time.Millisecond * 30)
				if v, ok := c.Get("a"); !ok || string(v) != "2" {
					t.Errorf("expected updated value=2, got=%v", v)
				}
			},
		},
		{
			name:     "janitor removes expired",
			capacity: 2,
			ttl:      time.Millisecond * 50,
			actions: func(c *LRUCache, t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				c.StartJanitor(ctx)

				c.Set("a", []byte("1"))
				time.Sleep(time.Millisecond * 60)

				c.cleanup()

				if _, ok := c.Get("a"); ok {
					t.Errorf("expected janitor cleanup to remove expired key")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewLRUCache(tt.capacity, tt.ttl)
			tt.actions(c, t)
		})
	}
}
