package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/san-pang/etcdv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"server/helloworld"
	"strconv"
	"syscall"
)

var (
	serv = flag.String("service", "", "service name")
	host = flag.String("host", "", "listening host")
	port = flag.Int("port", 0, "listening port")
	regAddr  = flag.String("register", "", "register etcd address")
)

func main()  {
	// 解析入参
	flag.Parse()
	// 开启监听端口
	lis, err := net.Listen("tcp", "0.0.0.0:" + strconv.Itoa(*port))
	if err != nil {
		log.Panic(err)
	}
	defer lis.Close()
	// 如果没有传入端口，或者传入的端口为0，会随机监听一个可用端口，需要获取实际监听的端口用于服务注册
	port = &lis.Addr().(*net.TCPAddr).Port
	reg := etcdv3.NewEtcdRegister(*regAddr, *serv, *host + ":" + strconv.Itoa(*port), 15)
	if err := reg.Register(); err != nil {
		log.Panic(err)
	}
	defer reg.UnRegister()
	// 创建grpc服务
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &server{})
	reflection.Register(s)
	// 接收退出信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		sig := <-ch
		log.Println("receive signal:", sig)
		// 服务取消注册
		reg.UnRegister()
		os.Exit(1)
	}()
	s.Serve(lis)
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Println("receive a hello greeter from : ", in.Name)
	return &helloworld.HelloReply{Message: fmt.Sprintf("%s, nice to meet you, i'm %s:%d", in.Name, *host, *port)}, nil
}