/*
Packages tun provides utilities for tun.

tun包提供 创建tun设备的方法，以及监听tun，将数据解析为tcp/udp数据的方法。

tun 工作在第三层 IP层上。

我们基本上抄了 xjasonlyu/tun2socks, 因此把GPL证书放在了本包的文件夹中

windows中,
需要从 https://www.wintun.net/ 下载 wintun.dll 放到vs可执行文件旁边
*/
package tun

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/e1732a364fed/v2ray_simple/netLayer"
	"github.com/e1732a364fed/v2ray_simple/netLayer/tun/device"
	"github.com/e1732a364fed/v2ray_simple/netLayer/tun/device/tun"
	"github.com/e1732a364fed/v2ray_simple/netLayer/tun/option"
	"github.com/e1732a364fed/v2ray_simple/utils"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func Open(name string) (device.Device, error) {
	return tun.Open(name, uint32(utils.MTU))
}

type StackCloser struct {
	*stack.Stack
}

// Close() and Wait()
func (sc *StackCloser) Close() error {
	sc.Stack.Close()
	sc.Stack.Wait() //这个会卡住
	return nil
}

func Listen(dev device.Device) (tcpChan chan netLayer.TCPRequestInfo, udpChan chan netLayer.UDPRequestInfo, closer io.Closer, err error) {

	tcpChan = make(chan netLayer.TCPRequestInfo)
	udpChan = make(chan netLayer.UDPRequestInfo)

	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
	})

	closer = &StackCloser{Stack: s}

	opts := []option.Option{option.WithDefault()}

	for _, opt := range opts {
		if err = opt(s); err != nil {
			return
		}
	}

	nicID := tcpip.NICID(s.UniqueID())

	if ex := s.CreateNICWithOptions(nicID, dev,
		stack.NICOptions{
			Disabled: false,
			// If no queueing discipline was specified
			// provide a stub implementation that just
			// delegates to the lower link endpoint.
			QDisc: nil,
		}); ex != nil {
		err = utils.ErrInErr{ErrDesc: ex.String()}
		return
	}

	const defaultWndSize = 0
	const maxConnAttempts int = 2048

	tcpForwarder := tcp.NewForwarder(s, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest) {
		var (
			wq  waiter.Queue
			ep  tcpip.Endpoint
			err tcpip.Error
			id  = r.ID()
		)

		// Perform a TCP three-way handshake.
		ep, err = r.CreateEndpoint(&wq)
		if err != nil {
			// RST: prevent potential half-open TCP connection leak.
			r.Complete(true)
			return
		}

		setSocketOptions(s, ep)

		tcpConn := gonet.NewTCPConn(&wq, ep)

		info := netLayer.TCPRequestInfo{
			Conn: tcpConn,

			//比较反直觉
			Target: netLayer.Addr{
				Network: "tcp",
				IP:      net.IP(id.LocalAddress),
				Port:    int(id.LocalPort),
			},
		}

		// log.Printf("forward tcp request %s:%d->%s:%d\n",
		// 	id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort)

		tcpChan <- info

		r.Complete(false)
	})
	s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

	udpForwarder := udp.NewForwarder(s, func(r *udp.ForwarderRequest) {
		var (
			wq waiter.Queue
			id = r.ID()
		)
		ep, err := r.CreateEndpoint(&wq)
		if err != nil {
			log.Printf("tun Err, udp forwarder request %s:%d->%s:%d: %\n",
				id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort, err)
			return
		}

		udpConn := gonet.NewUDPConn(s, &wq, ep)

		info := netLayer.UDPRequestInfo{
			MsgConn: &netLayer.MsgConnForPacketConn{PacketConn: udpConn},

			Target: netLayer.Addr{
				Network: "udp",
				IP:      net.IP(id.LocalAddress),
				Port:    int(id.LocalPort),
			},
		}

		udpChan <- info
	})
	s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)

	s.SetPromiscuousMode(nicID, true) //必须调用这个,否则tun什么也收不到
	s.SetSpoofing(nicID, true)

	s.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			NIC:         nicID,
		},
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         nicID,
		},
	})

	return
}

func setSocketOptions(s *stack.Stack, ep tcpip.Endpoint) tcpip.Error {
	{ /* TCP keepalive options */
		ep.SocketOptions().SetKeepAlive(true)

		const tcpKeepaliveIdle time.Duration = 60000000000

		idle := tcpip.KeepaliveIdleOption(tcpKeepaliveIdle)
		if err := ep.SetSockOpt(&idle); err != nil {
			return err
		}

		const tcpKeepaliveInterval time.Duration = 30000000000
		interval := tcpip.KeepaliveIntervalOption(tcpKeepaliveInterval)
		if err := ep.SetSockOpt(&interval); err != nil {
			return err
		}

		const tcpKeepaliveCount int = 9
		if err := ep.SetSockOptInt(tcpip.KeepaliveCountOption, tcpKeepaliveCount); err != nil {
			return err
		}
	}
	{ /* TCP recv/send buffer size */
		var ss tcpip.TCPSendBufferSizeRangeOption
		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &ss); err == nil {
			ep.SocketOptions().SetReceiveBufferSize(int64(ss.Default), false)
		}

		var rs tcpip.TCPReceiveBufferSizeRangeOption
		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &rs); err == nil {
			ep.SocketOptions().SetReceiveBufferSize(int64(rs.Default), false)
		}
	}
	return nil
}
