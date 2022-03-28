package grpc

import (
	"context"
	"net"
	sync "sync"
	"time"

	"github.com/hahahrfool/v2ray_simple/netLayer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clientconnMap   = make(map[netLayer.HashableAddr]ClientConn)
	clientconnMutex sync.RWMutex
)

type ClientConn *grpc.ClientConn

/*
调用过程

先用 GetEstablishedConnFor 看看有没有已存在的 clientConn

没有已存在的话，自己先拨号tcp，然后拨号tls，然后把tls连接 传递给 ClientHandshake, 生成一个 clientConn

然后把获取到的 clientConn传递给 DialNewSubConn, 获取可用的一条 grpc 连接

*/

//获取与 某grpc服务器的 已存在的grpc连接
func GetEstablishedConnFor(addr *netLayer.Addr) ClientConn {
	clientconnMutex.RLock()
	clieintconn := clientconnMap[addr.GetHashable()]
	clientconnMutex.RUnlock()

	if clieintconn == nil {
		return nil
	}

	if (*grpc.ClientConn)(clieintconn).GetState() != connectivity.Shutdown {
		return clieintconn
	}
	return nil
}

// ClientHandshake 在客户端被调用, 将一个普通连接升级为 grpc连接
//该 underlay一般为 tls连接。 addr为实际的远程地址，我们不从 underlay里获取addr,避免转换.
func ClientHandshake(underlay net.Conn, addr *netLayer.Addr) (ClientConn, error) {

	//v2ray的实现中用到了一个  globalDialerMap[dest], 可以利用现有连接,
	// 只有map里没有与对应目标远程地址的连接的时候才会拨号;
	// 这 应该就是一种mux的实现
	// 如果有之前播过的client的话，直接利用现有client进行 NewStream, 然后服务端的话, 实际上会获得到第二条连接;

	//也就是说，底层连接在客户端-服务端只用了一条，但是 服务端处理时却会抽象出 多条Stream连接进行处理

	grpc_clientConn, err := grpc.Dial("", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, addrStr string) (net.Conn, error) {
		return underlay, nil
	}), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  500 * time.Millisecond,
			Multiplier: 1.5,
			Jitter:     0.2,
			MaxDelay:   19 * time.Second,
		},
		MinConnectTimeout: 5 * time.Second,
	}))
	if err != nil {
		return nil, err
	}

	clientconnMutex.Lock()
	clientconnMap[addr.GetHashable()] = grpc_clientConn
	clientconnMutex.Unlock()

	return grpc_clientConn, nil

}

//在一个已存在的grpc连接中 进行新的子连接申请
func DialNewSubConn(path string, clientconn ClientConn, addr *netLayer.Addr) (net.Conn, error) {

	//不像服务端需要自己写一个实现StreamServer接口的结构, 我们Client端直接可以调用函数生成 StreamClient
	// 这也是grpc的特点, 客户端只负责 “调用“ ”service“，而具体的service的实现 是在服务端.

	streamClient := NewStreamClient((*grpc.ClientConn)(clientconn)).(StreamClient_withName)

	stream_TunClient, err := streamClient.Tun_withName(nil, path)
	if err != nil {
		clientconnMutex.Lock()
		delete(clientconnMap, addr.GetHashable())
		clientconnMutex.Unlock()
		return nil, err
	}
	return NewConn(stream_TunClient, nil), nil

}