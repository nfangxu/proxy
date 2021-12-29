package proxy

import (
	"fmt"
)

func NewFactory(configs map[string]*Config) *Factory {
	if configs == nil {
		configs = map[string]*Config{}
	}
	return &Factory{configs: configs, instances: map[string]*Proxy{}}
}

type Factory struct {
	configs   map[string]*Config
	instances map[string]*Proxy
}

func (f *Factory) Make(proxy string) (*Proxy, error) {
	if c, ok := f.instances[proxy]; ok {
		return c, nil
	}
	if c, ok := f.configs[proxy]; ok {
		p, e := NewProxy(c)
		if e != nil {
			return nil, fmt.Errorf("can not make proxy[%s] with error: %v", proxy, e)
		}
		f.instances[proxy] = p
		return p, nil
	}
	return nil, fmt.Errorf("unknown proxy named %s, place retry", proxy)
}
