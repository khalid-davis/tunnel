package udpmsg

import (
	"fmt"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/proxy/udp/udpmng"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog"
	"net"
)

// 云端server调用,然后由云端的tcp来write到与云端tcp client的连接上
func BackendHandler(msg *proto.StreamMsg) error {
	conn := context.GetContext().GetConn(msg.Topic)
	if conn == nil {
		klog.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s", msg.Topic, msg.Category, msg.Type)
		return fmt.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s ", msg.Topic, msg.Category, msg.Type)
	}
	conn.Send2Conn(msg)
	return nil
}

// 边端agent调用
func FrontendHandler(msg *proto.StreamMsg) error {
	c := context.GetContext().GetConn(msg.Topic)
	if c != nil {
		c.Send2Conn(msg)
		return nil
	}
	udp := udpmng.NewUDPConn(msg.Topic, msg.Addr, msg.Node, "agent")
	udp.Type = util.UDP_BACKEND
	udp.C.Send2Conn(msg)
	udpAddr, err := net.ResolveUDPAddr("udp", udp.FinalAddr)
	if err != nil {
		klog.Error("edge proxy resolve addr fail")
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		klog.Error("edge proxy connect fail")
		return err
	}
	udp.Conn = conn
	go udp.Read()
	go udp.Write()
	return nil
}

func ControlHandler(msg *proto.StreamMsg) error {
	conn := context.GetContext().GetConn(msg.Topic)
	if conn == nil {
		klog.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s", msg.Topic, msg.Category, msg.Type)
		return fmt.Errorf("trace_id = %s the stream module failed to distribute the side message module = %s type = %s ", msg.Topic, msg.Category, msg.Type)
	}
	conn.Send2Conn(msg)
	return nil
}