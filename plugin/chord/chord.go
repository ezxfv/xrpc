package chord

import (
	"encoding/json"

	chord "x.io/xrpc/app/chord/client"
	"x.io/xrpc/types"
)

func New(srvAddr, chordAddr string) *chordPlugin {
	return &chordPlugin{
		srvAddr: srvAddr,
		client:  chord.NewChordClient(chordAddr),
	}
}

type service struct {
	ServiceName string
	Methods     map[string]bool
	Endpoints   map[string]bool
}

type chordPlugin struct {
	srvAddr string
	client  *chord.ChordClient
}

func (c *chordPlugin) Start() error {
	return nil
}

func (c *chordPlugin) register(sd *types.ServiceDesc) error {
	serviceJson, err := c.client.Get(sd.ServiceName)
	if err != nil {
		return err
	}

	s := &service{}
	if len(serviceJson) > 0 {
		err = json.Unmarshal([]byte(serviceJson), s)
		if err != nil {
			return err
		}
		s.Endpoints[c.srvAddr] = true
	} else {
		s.ServiceName = sd.ServiceName
		s.Methods = map[string]bool{}
		for _, m := range sd.Methods {
			s.Methods[m.MethodName] = true
		}
		s.Endpoints = map[string]bool{c.srvAddr: true}
	}
	newJson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return c.client.Set(sd.ServiceName, string(newJson))
}

func (c *chordPlugin) RegisterService(sd *types.ServiceDesc, ss interface{}) error {
	return c.register(sd)
}

func (c *chordPlugin) RegisterCustomService(sd *types.ServiceDesc, ss interface{}, metadata string) error {
	return c.register(sd)
}

func (c *chordPlugin) RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	serviceJson, err := c.client.Get(serviceName)
	if err != nil {
		return err
	}

	s := &service{}
	if len(serviceJson) > 0 {
		err = json.Unmarshal([]byte(serviceJson), s)
		if err != nil {
			return err
		}
		s.Methods[fname] = true
		s.Endpoints[c.srvAddr] = true
	} else {
		s.ServiceName = serviceName
		s.Methods = map[string]bool{}
		s.Methods[fname] = true
		s.Endpoints = map[string]bool{c.srvAddr: true}
	}
	newJson, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return c.client.Set(serviceName, string(newJson))
}

func (c *chordPlugin) Unregister(serviceName string) error {
	return c.client.Del(serviceName)
}
