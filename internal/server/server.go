package server

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	"review-server/internal/conf"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, ReviewRegister)

func ReviewRegister(conf *conf.Register) registry.Registrar {
	//c := api.DefaultConfig()
	c := &api.Config{
		Address: conf.Consul.Addr,
		Scheme:  conf.Consul.Scheme,
	}
	client, err := api.NewClient(c)
	if err != nil {
		panic("register review-service has error")
	}
	return consul.New(client)
}
