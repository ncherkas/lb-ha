package store

import (
	"cmp"
	"log"
	"slices"
	"sync"
)

// StoreEntry with the fields store in memory store
type StoreEntry struct {
	Key       string
	Value     string
	Timestamp int64
}

// EntryWriter is used to abstract away DumpAll writing mechanism
type EntryWriter interface {
	write(skv *StoreEntry)
}

// EntryWriter implementation that writes entries into the log
type EntryLogger struct {
}

func (l *EntryLogger) write(skv *StoreEntry) {
	log.Printf("[DEBUG] %s -> %s @%d\n", skv.Key, skv.Value, skv.Timestamp)
}

// MemStore implementation build around sync.Map
type MemStore struct {
	m            sync.Map
	GetAllwriter EntryWriter
}

func New() *MemStore {
	return &MemStore{GetAllwriter: new(EntryLogger)}
}

func (s *MemStore) Add(key, val string, timestamp int64) {
	s.m.Store(key, &StoreEntry{Key: key, Value: val, Timestamp: timestamp})
}

func (s *MemStore) Delete(key string) {
	s.m.Delete(key)
}

func (s *MemStore) Get(key string) (found bool, val string, timestamp int64) {
	if v, ok := s.m.Load(key); ok {
		se := v.(*StoreEntry)
		found, val, timestamp = true, se.Value, se.Timestamp
	}
	return
}

func (s *MemStore) DumpAll() {
	entries := []*StoreEntry{}
	s.m.Range(func(key, value any) bool {
		entries = append(entries, value.(*StoreEntry))
		return true
	})
	slices.SortFunc(entries, func(a, b *StoreEntry) int {
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})
	for i := range entries {
		skv := entries[i]
		s.GetAllwriter.write(skv)
		entries[i] = nil // Cleaning up them for quicker GC
	}
	s = nil
}
