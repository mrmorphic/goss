package cache

import (
	"testing"
	"time"
)

func TestSaveCache(t *testing.T) {
	key := "Key"
	value := "Value"

	// test that the cache key is not set to start with
	v := Get(key)
	if v != nil {
		t.Errorf("Did not expect cache key '%s' to be set, but has value '%s'", key, v)
	}

	Store(key, value, time.Second*30)

	v = Get(key)
	if v == nil {
		t.Errorf("Expected cache key '%s' to have a value, but returned nil", key)
	} else if v.(string) != value {
		t.Errorf("Expected cache key '%s' to have value '%s', but got '%s'", key, value, v)
	}
}

func TestExpiry(t *testing.T) {
	key := "Key2"
	value := "Value2"
	seconds := time.Duration(1)

	// set a key with short expiry
	Store(key, value, time.Second*seconds)

	// test we can retrieve it immediately
	v := Get(key)
	if v == nil || v.(string) != value {
		t.Errorf("Expected cache key '%s' to have value '%s', but got '%s'", key, value, v)
	}

	// wait past expiry, giving a couple of extra seconds
	time.Sleep(time.Second * (seconds + 1))

	// test it's gone
	v = Get(key)
	if v != nil {
		t.Errorf("Did not expect cache key '%s' to be still set after expiry, but has value '%s'", key, v)
	}
}
