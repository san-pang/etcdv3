package main

import (
	"client/helloworld"
	"context"
	"flag"
	"github.com/san-pang/etcdv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

var (
	svc = flag.String("service", "hello_service", "service name")
	reg = flag.String("register", "127.0.0.1:2379", "register etcd address")
)

func main()  {
	flag.Parse()
	r := etcdv3.NewResolver(*reg, *svc)
	resolver.Register(r)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://authority/"+*svc, grpc.WithBalancerName(roundrobin.Name), grpc.WithInsecure(), grpc.WithBlock())
	cancel()
	if err != nil {
		log.Panic(err)
	}
	client := helloworld.NewGreeterClient(conn)
	for {
		resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Name: "san-pang"})
		if err != nil {
			log.Println(err)
		} else {
			log.Println("receive response:", resp.Message)
		}
		<-time.After(time.Second)
	}
}
