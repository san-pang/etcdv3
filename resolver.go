package etcdv3

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	etcd "go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"strings"
	"time"
)

const schema = "san-pang"

type etcdResolver struct {
	etcdAddr  	string
	serviceName string
	cc      	resolver.ClientConn
	cli			*etcd.Client
}

func NewResolver(etcdAddr, serviceName string) resolver.Builder {
	return &etcdResolver{
		etcdAddr:    etcdAddr,
		serviceName: serviceName,
		cc:			 nil,
		cli:         nil,
	}
}

func (e *etcdResolver) Scheme() string {
	return schema
}

func (e *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	var err error
	e.cli, err = etcd.New(etcd.Config{
		Endpoints: strings.Split(e.etcdAddr, ";"),
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		return nil, fmt.Errorf("etcd client create failed: %v", err)
	}
	e.cc = cc
	go e.watch("/" + schema + "/" + e.serviceName + "/")
	return e, nil
}

func (e *etcdResolver) ResolveNow(rn resolver.ResolveNowOption) {

}

func (e *etcdResolver) Close() {

}

func (e *etcdResolver) watch(prefix string) {
	addrDict := make(map[string]resolver.Address)

	update := func() {
		addrList := make([]resolver.Address, 0, len(addrDict))
		for _, v := range addrDict {
			addrList = append(addrList, v)
		}
		e.cc.UpdateState(resolver.State{Addresses: addrList})
	}

	resp, err := e.cli.Get(context.Background(), prefix, etcd.WithPrefix())
	if err == nil {
		for i := range resp.Kvs {
			addrDict[string(resp.Kvs[i].Value)] = resolver.Address{Addr: string(resp.Kvs[i].Value)}
		}
	}

	update()

	rch := e.cli.Watch(context.Background(), prefix, etcd.WithPrefix(), etcd.WithPrevKV())
	for n := range rch {
		for _, ev := range n.Events {
			switch ev.Type {
			case mvccpb.PUT:
				addrDict[string(ev.Kv.Key)] = resolver.Address{Addr: string(ev.Kv.Value)}
			case mvccpb.DELETE:
				delete(addrDict, string(ev.PrevKv.Key))
			}
		}
		update()
	}
}