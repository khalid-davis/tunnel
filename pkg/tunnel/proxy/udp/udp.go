package udp

import (
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/model"
	"github.com/superedge/superedge/pkg/tunnel/proxy/udp/udpmng"
	"github.com/superedge/superedge/pkg/tunnel/proxy/udp/udpmsg"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog"
	"net"
	"time"
)

type UDPProxy struct {
}

func (udp *UDPProxy) Register(ctx *context.Context) {
	ctx.AddModule(udp.Name())
}

func (udp *UDPProxy) Name() string {
	return util.UDP
}

func (udp *UDPProxy) Start(mode string) {
	// server处理函数
	context.GetContext().RegisterHandler(util.UDP_BACKEND, udp.Name(), udpmsg.BackendHandler)
	// agent处理函数
	context.GetContext().RegisterHandler(util.UDP_FRONTEND, udp.Name(), udpmsg.FrontendHandler)
	context.GetContext().RegisterHandler(util.UDP_CONTROL, udp.Name(), udpmsg.ControlHandler)
	// cloud
	// 在linux跨节点的时候要使用0.0.0.0，不能使用127
	//front := "0.0.0.0:6442"
	//front := "127.0.0.1:6442"
	//backend := "192.168.0.98:8285"
	//backend := "127.0.0.1:8286"
	// edge
	//front := "0.0.0.0:6443"
	front := "127.0.0.1:6443"
	//backend := "192.168.0.226:8285"
	backend := "127.0.0.1:8285"
	go func(front, backend string) {
		udpAddr, err := net.ResolveUDPAddr("udp", front)
		ln, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			klog.Errorf("cloud proxy start %s fail ,error = %s", front, err)
			return
		}
		defer ln.Close()
		klog.Infof("the udp server of the cloud tunnel listen on %s\n", front)
		// 这里会不会有问题，消息怎么分别，在tcp里面每个请示会有一个conn，但是在这个里面呢？，UDP并没有Accept函数
		// 后面设计的时候需要把readFromUDP那里得到的addr直接放到一个新的协程里面去处理并等待响应，因为客户端的回写地址是不一样的，要做响应分发的
		var node string
		for {
			nodes := context.GetContext().GetNodes()
			if len(nodes) == 0 {
				klog.Errorf("len(nodes)==0")
				time.Sleep(5 * time.Second)
				continue
			}
			node = nodes[0]
			break
		}
		klog.Infof("nodeName: ", node)
		uuid := uuid.NewV4().String()
		fp := udpmng.NewUDPConn(uuid, backend, node)
		fp.Conn = ln
		fp.Type = util.UDP_FRONTEND
		// 在这种模式下面，是不需要监听响应的，响应是以同样的方式，对方为client，自己为server进行返回的
		// 这里只需要做单边就可以了，专门负责监听代理到的消息并把消息通过grpc传送到对端。与FrontendHandler配套使用
		//go fp.Write()
		go fp.Read()
		<-fp.StopChan
	}(front, backend)
}

func (udp *UDPProxy) CleanUp() {
	context.GetContext().RemoveModule(udp.Name())
}

func InitUDP() {
	model.Register(&UDPProxy{})
	klog.Infof("init module: %s success!", util.UDP)
}