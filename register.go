package etcdv3

import (
	"context"
	"errors"
	"fmt"
	etcd "go.etcd.io/etcd/clientv3"
	"log"
	"strings"
)

type EtcdRegister struct {
	etcdAddr string
	serviceName string
	serviceAddr string
	ttl int64
	cli *etcd.Client
	key string
}

func NewEtcdRegister(etcdAddr, serviceName, serviceAddr string, ttl int64) *EtcdRegister {
	return &EtcdRegister{
		etcdAddr:    etcdAddr,
		serviceName: serviceName,
		serviceAddr: serviceAddr,
		ttl:         ttl,
		cli:         nil,
		key: 		 "/"+schema+"/"+serviceName+"/"+serviceAddr,
	}
}

func (e *EtcdRegister)Register() error {
	/*
	服务注册实现
	etcdAddr: etcd的地址，存在多个时以分号分割，例如127.0.0.1:2379;127.0.0.1:2389
	serviceName: 注册的服务名称
	serviceAddr: 服务监听地址，例如127.0.0.1:8800
	ttl: ttl值
	 */
	var err error
	if e.cli != nil {
		return errors.New("dulplicate register")
	}
	e.cli, err = etcd.New(etcd.Config{
		Endpoints: strings.Split(e.etcdAddr, ";"),
	})
	if err != nil {
		return fmt.Errorf("etcd connect failed: %v", err)
	}

	log.Println("start grant to etcd......")
	resp, err := e.cli.Grant(context.TODO(), e.ttl)
	if err != nil {
		return fmt.Errorf("ctreate etcd lease failed: %v", err)
	}
	log.Println("grant success!")

	log.Println("start register key to etcd......")
	if _, err := e.cli.Put(context.TODO(), e.key, e.serviceAddr, etcd.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("register service with ttl into etcd failed: %v", err)
	}
	log.Println("register key success!")

	log.Println("start keep alive......")
	if _, err := e.cli.KeepAlive(context.TODO(), resp.ID); err != nil {
		return fmt.Errorf("service keepAlive with ttl failed: %v", err)
	}
	log.Println("keep alive success!")
	return nil
}

func (e *EtcdRegister)UnRegister() error {
	if e.cli != nil {
		_, err := e.cli.Delete(context.Background(), e.key)
		return err
	}
	return nil
}