package dns

import (
	"github.com/miekg/dns"
	"sync"
	"time"
)

type Cache struct {
	cache map[msgKey]msgExpire
	lock  sync.RWMutex
}

func NewCache() *Cache {
	c := &Cache{
		cache: make(map[msgKey]msgExpire),
		lock:  sync.RWMutex{},
	}
	go func() {
		time.Sleep(10 * time.Minute)
		c.cleanUp()
	}()
	return c
}

type msgKey struct {
	domain, ecs string
	typ         uint16
}

type msgExpire struct {
	msg    *dns.Msg
	expire time.Time
}

func (c *Cache) Put(domain string, typ uint16, ecs string, m *dns.Msg) {
	var expire time.Time
	if len(m.Answer) > 0 {
		expire = time.Now().Add(time.Duration(m.Answer[0].Header().Ttl) * time.Second)
	} else {
		expire = time.Now().Add(10 * time.Minute)
	}

	c.lock.Lock()
	c.cache[msgKey{domain: domain, typ: typ, ecs: ecs}] = msgExpire{m, expire}
	c.lock.Unlock()
}

func (c *Cache) Get(domain string, typ uint16, ecs string) (m *dns.Msg) {
	key := msgKey{domain: domain, typ: typ, ecs: ecs}
	c.lock.RLock()
	me, ok := c.cache[key]
	c.lock.RUnlock()
	if ok && me.expire.After(time.Now()) {
		return me.msg
	}
	return nil
}

func (c *Cache) cleanUp() {
	del := make([]msgKey, 0, len(c.cache))
	now := time.Now()
	c.lock.RLock()
	for k, v := range c.cache {
		if v.expire.Before(now) {
			del = append(del, k)
		}
	}
	c.lock.RUnlock()

	if len(del) > 0 {
		c.lock.Lock()
		for _, k := range del {
			delete(c.cache, k)
		}
		c.lock.Unlock()
	}
}
