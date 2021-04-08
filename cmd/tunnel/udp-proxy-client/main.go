package main

import (
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/model"
	"github.com/superedge/superedge/pkg/tunnel/proxy/stream"
	"github.com/superedge/superedge/pkg/tunnel/proxy/udp"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"log"
	"os"
)

func main()  {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "./conf/edge_mode.toml")
	if err != nil {
		log.Printf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	model.InitModules(util.EDGE)
	udp.InitUDP()
	stream.InitStream(util.EDGE)
	model.LoadModules(util.EDGE)
	model.ShutDown()
}