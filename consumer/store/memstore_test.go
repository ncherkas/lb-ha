package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEntryWriter struct {
	mock.Mock
}

func (m *MockEntryWriter) write(skv *StoreEntry) {
	m.Called(skv)
}

func TestMemStore_Add(t *testing.T) {
	store := New()
	ts := time.Now().Unix()
	store.Add("key1", "value1", ts)
	val, ok := store.m.Load("key1")
	assert.True(t, ok)
	se := val.(*StoreEntry)
	assert.Equal(t, "value1", se.Value)
	assert.Equal(t, ts, se.Timestamp)
}

func TestMemStore_Delete(t *testing.T) {
	store := New()
	store.Add("key1", "value1", time.Now().Unix())
	store.Delete("key1")
	_, ok := store.m.Load("key1")
	assert.False(t, ok)
}

func TestMemStore_Get(t *testing.T) {
	store := New()
	store.Add("key1", "value1", 1)
	store.Add("key2", "value2", 2)

	ok, val, ts := store.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)
	assert.Equal(t, int64(1), ts)

	ok, val, ts = store.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)
	assert.Equal(t, int64(2), ts)

	ok, val, ts = store.Get("key3")
	assert.False(t, ok)
	assert.Equal(t, "", val)
	assert.Equal(t, int64(0), ts)
}

func TestMemStore_DumpAll(t *testing.T) {
	mockWriter := new(MockEntryWriter)
	store := &MemStore{DumpAllWriter: mockWriter}

	// Adding entries to the store
	store.Add("key1", "value1", 1)
	store.Add("key2", "value2", 2)
	store.Add("key3", "value3", 3)
	store.Add("key4", "value4", 4)
	store.Add("key5", "value5", 5)

	// Expected calls to mock writer
	mockWriter.On("write", &StoreEntry{Key: "key1", Value: "value1", Timestamp: 1}).Once()
	mockWriter.On("write", &StoreEntry{Key: "key2", Value: "value2", Timestamp: 2}).Once()
	mockWriter.On("write", &StoreEntry{Key: "key3", Value: "value3", Timestamp: 3}).Once()
	mockWriter.On("write", &StoreEntry{Key: "key4", Value: "value4", Timestamp: 4}).Once()
	mockWriter.On("write", &StoreEntry{Key: "key5", Value: "value5", Timestamp: 5}).Once()

	store.DumpAll()

	mockWriter.AssertExpectations(t)
}
