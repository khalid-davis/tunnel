package udpmng

import (
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io"
	"k8s.io/klog"
	"net"
	"sync"
)

type UDPConn struct {
	Conn       *net.UDPConn
	role       string // server、agent
	uid        string
	StopChan   chan struct{}
	Type       string
	C          context.Conn
	n          context.Node
	FinalAddr  string // agent用于建立数据通道的地址 （最终一站）
	SourceAddr *net.UDPAddr // server 回传UDP报文到一开始使用的地址（第一站）
	once       sync.Once
}

func NewUDPConn(uuid, addr, node, role string) *UDPConn {
	udp := &UDPConn{
		uid:       uuid,
		StopChan:  make(chan struct{}, 1),
		C:         context.GetContext().AddConn(uuid),
		FinalAddr: addr,
		role:      role,
	}

	if nd := context.GetContext().GetNode(node); nd != nil {
		udp.n = nd
	} else {
		udp.n = context.GetContext().AddNode(node)
	}
	udp.n.BindNode(uuid)
	return udp
}

func (udp *UDPConn) Write() {
	running := true
	for running {
		select {
		case msg := <-udp.C.ConnRecv():
			if msg.Type == util.UDP_CONTROL {
				udp.cleanUp()
				break
			}
			var err error
			// 通过DialUDP方式创建的UDP是已连接的，请示会记录remote的信息，而服务端通过 listenUDP创建的连接是未连接的，在发送和请示数据时需要填充remote信息
			if udp.role == "server" {
				// 回写
				_, err = udp.Conn.WriteToUDP(msg.Data, udp.SourceAddr)
			} else {
				_, err = udp.Conn.Write(msg.Data)
			}
			if err != nil {
				klog.Errorf("write conn fail err = %v", err)
				udp.cleanUp()
				break
			}
		case <-udp.StopChan:
			klog.Info("disconnect udp and stop sending")
			udp.Conn.Close()
			udp.closeOpposite()
			running = false
		}
	}
}


func (udp *UDPConn) Read() {
	running := true
	for running {
		select {
		case <-udp.StopChan:
			klog.Info("Disconnect udp and stop receiving")
			udp.Conn.Close()
			running = false
		default:
			size := 32 * 1024
			if l, ok := interface{}(udp.Conn).(*io.LimitedReader); ok && int64(size) > l.N {
				if l.N < 1 {
					size = 1
				} else {
					size = int(l.N)
				}
			}
			buf := make([]byte, size)
			var n int
			var err error
			if udp.role == "server" {
				var addr *net.UDPAddr
				n, addr, err = udp.Conn.ReadFromUDP(buf)
				udp.SourceAddr = addr
			} else {
				n, err = udp.Conn.Read(buf)
			}
			if err != nil {
				klog.Errorf("read udp failed, err = %s ", err)
				udp.cleanUp()
				break
			}
			fmt.Println("data: ", string(buf))
			udp.n.Send2Node(&proto.StreamMsg{
				Node:     udp.n.GetName(),
				Category: util.UDP,
				Type:     udp.Type,
				Topic:    udp.uid,
				Data:     buf[0:n],
				Addr:     udp.FinalAddr,
			})
		}
	}
}


func (udp *UDPConn) cleanUp() {
	udp.StopChan <- struct{}{}
}


func (udp *UDPConn) closeOpposite() {
	udp.once.Do(func() {
		udp.n.Send2Node(&proto.StreamMsg{
			Category:             util.TCP,
			Type:                 util.TCP_CONTROL,
			Topic:                udp.uid,
			Data:                 []byte{},
			Addr:                 udp.FinalAddr,
		})
	})
}