package main

import (
	"fmt"
	"sync"
)

// ProxyCollection is a collection of proxies. It's the interface for anything
// to add and remove proxies from the toxiproxy instance. It's responsibilty is
// to maintain the integrity of the proxy set, by guarding for things such as
// duplicate names.
type ProxyCollection struct {
	sync.RWMutex

	proxies map[string]*Proxy
}

func NewProxyCollection() *ProxyCollection {
	return &ProxyCollection{
		proxies: make(map[string]*Proxy),
	}
}

func (collection *ProxyCollection) Add(proxy *Proxy) error {
	collection.Lock()
	defer collection.Unlock()

	if _, exists := collection.proxies[proxy.Name]; exists {
		return fmt.Errorf("Proxy with name %s already exists", proxy.Name)
	}

	collection.proxies[proxy.Name] = proxy

	return nil
}

// Because maps aren't thread-safe, the lock is only valid for the duration of the passed in block
func (collection *ProxyCollection) Proxies(block func(map[string]*Proxy) error) error {
	collection.RLock()
	defer collection.RUnlock()

	return block(collection.proxies)
}

func (collection *ProxyCollection) Get(name string) (*Proxy, error) {
	collection.RLock()
	defer collection.RUnlock()

	return collection.getByName(name)
}

func (collection *ProxyCollection) Remove(name string) error {
	collection.Lock()
	defer collection.Unlock()

	proxy, err := collection.getByName(name)
	if err != nil {
		return err
	}
	proxy.Stop()

	delete(collection.proxies, proxy.Name)
	return nil
}

func (collection *ProxyCollection) Clear() error {
	collection.Lock()
	defer collection.Unlock()

	for _, proxy := range collection.proxies {
		proxy.Stop()

		delete(collection.proxies, proxy.Name)
	}

	return nil
}

// getByName returns a proxy by its name. Its used from #remove and #get.
// It assumes the lock has already been acquired.
func (collection *ProxyCollection) getByName(name string) (*Proxy, error) {
	proxy, exists := collection.proxies[name]
	if !exists {
		return nil, fmt.Errorf("Proxy with name %s doesn't exist", name)
	}
	return proxy, nil
}
