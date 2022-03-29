package synchronization

import "sync"

type ThreadSafeMap struct {
	sync.RWMutex
	m map[string]string
}

func (t *ThreadSafeMap) Add(key, value string) {
	t.Lock()
	defer t.Unlock()
	t.m[key] = value
}
func (t *ThreadSafeMap) Remove(key string) {
	t.Lock()
	defer t.Unlock()
	delete(t.m, key)
}

func (t *ThreadSafeMap) Get(key string) string {
	t.RLock()
	defer t.RUnlock()
	return t.m[key]
}
