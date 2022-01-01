package webdav

import (
	"context"
	dav "golang.org/x/net/webdav"
	"log"
	"net"
	"net/http"
	"sni/util"
	"sni/util/env"
	"time"
)

func StartHttpServer() {
	// Parse env vars:
	disabled := env.GetOrDefault("SNI_WEBDAV_DISABLE", "0")
	if util.IsTruthy(disabled) {
		log.Printf("webdav: server disabled due to env var %s=%s\n", "SNI_WEBDAV_DISABLE", disabled)
		return
	}

	listenAddr := env.GetOrDefault("SNI_WEBDAV_LISTEN_ADDR", "0.0.0.0:23064")
	go listenHttp(listenAddr)
}

func listenHttp(listenAddr string) {
	var err error
	var lis net.Listener

	//var du *url.URL
	//du, err = url.Parse("fxpakpro://./dev/cu.usbmodemDEMO000000001")
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//var dev snes.AutoCloseableDevice
	//_, dev, err = snes.DeviceByUri(du)
	//if err != nil {
	//	log.Println(err)
	//}

	adapterFileSystem := &AdapterFileSystem{
		//fs: dev,
	}

	mux := http.NewServeMux()
	mux.Handle("/", &dav.Handler{
		Prefix:     "/fxpakpro/0/",
		FileSystem: adapterFileSystem,
		LockSystem: dav.NewMemLS(),
		Logger: func(req *http.Request, err error) {
			if err != nil {
				log.Printf("%s %s: %v\n", req.Method, req.URL.String(), err)
				return
			}
			log.Printf("%s %s {%+v}\n", req.Method, req.URL.String(), req.Header)
		},
	})

	// attempt to start the webdav server:
	count := 0
	lc := &net.ListenConfig{Control: util.ReusePortControl}
	for {
		lis, err = lc.Listen(context.Background(), "tcp", listenAddr)
		if err == nil {
			break
		}

		if count == 0 {
			log.Printf("webdav: failed to listen on %s: %v\n", listenAddr, err)
		}
		count++
		if count >= 30 {
			count = 0
		}

		time.Sleep(time.Second)
	}

	log.Printf("webdav: listening on %s\n", listenAddr)
	err = http.Serve(lis, mux)
	log.Printf("webdav: exit listenHttp: %v\n", err)
}
