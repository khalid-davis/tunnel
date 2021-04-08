package udp

import (
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/conf"
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
	if mode == util.CLOUD {
		// 先直接利用这个tcp字段
		for front, backend := range conf.TunnelConf.TunnlMode.Cloud.Tcp {
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
				uuid := uuid.NewV4().String()
				fp := udpmng.NewUDPConn(uuid, backend, node, "server")
				fp.Conn = ln
				fp.Type = util.UDP_FRONTEND
				go fp.Write()
				go fp.Read()
				<-fp.StopChan
			}(front, backend)
		}
	}
}

func (udp *UDPProxy) CleanUp() {
	context.GetContext().RemoveModule(udp.Name())
}

func InitUDP() {
	model.Register(&UDPProxy{})
	klog.Infof("init module: %s success!", util.UDP)
}