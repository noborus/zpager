package biomap

import "sync"

type Map[k comparable, v comparable] struct {
	s        sync.RWMutex
	Forward  map[k]v
	Backward map[v]k
}

func NewMap[k comparable, v comparable]() *Map[k, v] {
	return &Map[k, v]{
		Forward:  make(map[k]v),
		Backward: make(map[v]k),
	}
}

func (m *Map[k, v]) Store(key k, value v) {
	m.s.Lock()
	defer m.s.Unlock()
	m.Forward[key] = value
	m.Backward[value] = key
}

func (m *Map[k, v]) LoadForward(key k) (value v, ok bool) {
	m.s.RLock()
	defer m.s.RUnlock()
	value, ok = m.Forward[key]
	return
}

func (m *Map[k, v]) LoadBackward(value v) (key k, ok bool) {
	m.s.RLock()
	defer m.s.RUnlock()
	key, ok = m.Backward[value]
	return
}

func (m *Map[k, v]) DeleteForward(key k) {
	m.s.Lock()
	defer m.s.Unlock()
	value, ok := m.Forward[key]
	if !ok {
		return
	}
	delete(m.Forward, key)
	delete(m.Backward, value)
}

func (m *Map[k, v]) DeleteBackward(value v) {
	m.s.Lock()
	defer m.s.Unlock()
	key, ok := m.Backward[value]
	if !ok {
		return
	}
	delete(m.Forward, key)
	delete(m.Backward, value)
}
