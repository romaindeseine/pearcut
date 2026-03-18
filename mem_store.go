package pearcut

import (
	"sync"
)

type MemStore struct {
	mu    sync.RWMutex
	cache map[string]Experiment
}

func NewMemStore(experiments []Experiment) *MemStore {
	cache := make(map[string]Experiment, len(experiments))
	for _, exp := range experiments {
		cache[exp.Slug] = exp
	}
	return &MemStore{cache: cache}
}

func (m *MemStore) Get(slug string) (Experiment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exp, ok := m.cache[slug]
	if !ok {
		return Experiment{}, ErrExperimentNotFound
	}
	return exp, nil
}

func (m *MemStore) List(slugs []string, status ExperimentStatus) ([]Experiment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []Experiment

	if len(slugs) > 0 {
		for _, slug := range slugs {
			if exp, ok := m.cache[slug]; ok && exp.Status == status {
				results = append(results, exp)
			}
		}
	} else {
		for _, exp := range m.cache {
			if exp.Status == status {
				results = append(results, exp)
			}
		}
	}

	return results, nil
}

func (m *MemStore) Set(exp Experiment) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache[exp.Slug] = exp
}

func (m *MemStore) Delete(slug string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, slug)
}
