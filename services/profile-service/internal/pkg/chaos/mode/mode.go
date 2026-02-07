package mode

import "sync"

type Mode string

const (
	OK        Mode = "ok"
	Error     Mode = "error"
	RareError Mode = "rare_error"
	Slow      Mode = "slow"
	Flaky     Mode = "flaky"
	Overload  Mode = "overload"
)

type Store struct {
	mu      sync.RWMutex
	global  Mode
	methods map[string]Mode
}

func NewStore() *Store {
	return &Store{
		global:  OK,
		methods: make(map[string]Mode),
	}
}

func (s *Store) SetGlobal(m Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.global = m
}

func (s *Store) SetForMethod(method string, m Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[method] = m
}

func (s *Store) Get(method string) Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if m, ok := s.methods[method]; ok {
		return m
	}
	return s.global
}
