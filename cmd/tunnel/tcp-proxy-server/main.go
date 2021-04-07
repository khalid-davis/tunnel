package main

import (
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/model"
	"github.com/superedge/superedge/pkg/tunnel/proxy/stream"
	"github.com/superedge/superedge/pkg/tunnel/proxy/tcp"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"log"
)

func main() {
	err := conf.InitConf(util.CLOUD, "./conf/cloud_mode.toml")
	if err != nil {
		log.Printf("failed to initialize stream server configuration file err = %v", err)
		return
	}
	model.InitModules(util.CLOUD)
	tcp.InitTcp()
	stream.InitStream(util.CLOUD)
	model.LoadModules(util.CLOUD)
	model.ShutDown()
}

