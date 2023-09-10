package main

import "sync"

type logMetadata struct {
	mu          sync.RWMutex
	Term        uint64
	Index       uint64
	CommitIndex uint64
}

func (m *logMetadata) GetTerm() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Term
}

func (m *logMetadata) GetIndex() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Index
}

func (m *logMetadata) Set(term, index uint64) {
	m.mu.Lock()
	m.Term = term
	m.Index = index
	m.mu.Unlock()
}

func (m *logMetadata) GetCommitIndex() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CommitIndex
}

func (m *logMetadata) SetCommitIndex(commitIndex uint64) {
	m.mu.Lock()
	m.CommitIndex = commitIndex
	m.mu.Unlock()
}
