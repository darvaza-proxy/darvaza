/*
gpp-server is a server application support proxy, file serve  and https, http2.

Usage

use this command to generate a configure file and edit it
    $ gpp-server -dumpflags > server.ini


run it with
    $ gpp-server -config server.ini

use this command to show help message
    $ gpp-server -h

http2 is not enabled if you do not provide the tls certificate and private key file.
*/
package main

import (
	//"crypto/tls"
	"flag"
	"fmt"
	//"github.com/fangdingjun/gpp"
	"io/ioutil"
	// "github.com/fangdingjun/net/http2"
	//"github.com/gorilla/mux"
	//"github.com/vharitonsky/iniflags"
	//"log"
	//"net"
	//"encoding/base64"
	"github.com/fangdingjun/gpp/util"
	//"net/http"
	"os"
	//"strings"
	"encoding/json"
	"time"
)

// Router is a global router for http.Server
//var Router *mux.Router
var cfg CFG

func main() {
	//var ssl_cert tls.Certificate
	//var listener, listener1 net.Listener
	//var err error
	util.DialTimeout = 2 * time.Second
	//http2.VerboseLogs = false
	//var srv, srv1 http.Server
	var cfgFile string

	flag.StringVar(&cfgFile, "config", "", "config file")
	flag.Parse()

	buf, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		fmt.Printf("read configure file failed: %s\n", err.Error())
		os.Exit(-1)
	}

	err = json.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Printf("parse configure failed: %s", err)
		os.Exit(-1)
	}

	//Router = mux.NewRouter()

	initRouters()
	initListeners()

	// make a delay to make sure socket create success before drop privilege
	select {
	case <-time.After(2 * time.Second):
	}

	err = util.DropPrivilege(cfg.User, cfg.Group)
	if err != nil {
		fmt.Println(err)
	}

	select {}
}
