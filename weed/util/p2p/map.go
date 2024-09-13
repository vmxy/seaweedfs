package p2p

type KV[K, V any] struct {
	K K
	V V
}
type Map[K string | uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64, V any] struct {
	reads     map[K]V
	writes    map[K]V
	readLock  MicroLock
	writeLock MicroLock
}

func NewMap[K string | uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64, V any]() Map[K, V] {
	readLock := NewMicroLock()
	writeLock := NewMicroLock()
	m := Map[K, V]{
		reads:     make(map[K]V),
		writes:    make(map[K]V),
		readLock:  readLock,
		writeLock: writeLock,
	}
	return m
}
func (m *Map[K, V]) Get(key K) (val V, find bool) {
	var v V
	var ok bool
	if m.readLock.IsLock() {
		m.readLock.Lock()
		v, ok = m.reads[key]
		m.readLock.UnLock()
	} else {
		v, ok = m.reads[key]
	}
	if !ok {
		m.writeLock.Lock()
		v, ok = m.writes[key]
		m.writeLock.UnLock()
		if ok {
			m.readLock.Lock()
			m.reads[key] = v
			m.readLock.UnLock()
			return v, ok
		}
	}
	return v, ok
}
func (m *Map[K, V]) Set(key K, val V) {
	m.writeLock.Safe(func() { m.writes[key] = val })
}
func (m *Map[K, V]) Delete(key K) {
	m.writeLock.Safe(func() { delete(m.writes, key) })
	m.readLock.Safe(func() { delete(m.reads, key) })
}
func (m *Map[K, V]) Clear() {
	m.writeLock.Safe(func() { clear(m.writes) })
	m.readLock.Safe(func() { clear(m.reads) })
}
func (m *Map[K, V]) Size() int {
	return len(m.writes)
}
func (m *Map[K, V]) Range(handle func(k K, v V)) {
	m.writeLock.Safe(func() {
		for k, v := range m.writes {
			handle(k, v)
		}
	})
}
