package main

import (
	"flag"
	"log"
	"net/http"
	"sni/cmd/sni/appversion"
	"sni/cmd/sni/config"
	"sni/cmd/sni/logging"
	"sni/cmd/sni/tray"
	"sni/devices/snes/drivers/emunw"
	"sni/devices/snes/drivers/fxpakpro"
	"sni/devices/snes/drivers/luabridge"
	"sni/devices/snes/drivers/mock"
	"sni/devices/snes/drivers/retroarch"
	"sni/services/grpcimpl"
	"sni/services/usb2snes"
)

import _ "net/http/pprof"

// build variables set via ldflags by `go build -ldflags="-X 'main.version=v1.0.0'"`:
var (
	version string = "v0.0.0"
	commit  string = "dirty"
	date    string = "2021-05-03T00:17:00Z"
	builtBy string = "go"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "start pprof profiler on addr:port")
)

func main() {
	// make the version info public in the appversion package because the main package cannot be imported:
	appversion.Init(
		version,
		commit,
		date,
		builtBy,
	)

	// initialize tray, i.e. the Console window functionality:
	var err error
	err = tray.Init()
	if err != nil {
		log.Fatalln(err)
		return
	}

	// initialize logging subsystem:
	logging.Init()

	flag.Parse()
	if *cpuprofile != "" {
		go func() {
			// "localhost:6060"
			// /debug/pprof/
			log.Println(http.ListenAndServe(*cpuprofile, nil))
		}()
	}

	// load configuration:
	config.Load()

	// explicitly initialize all the drivers:
	fxpakpro.DriverInit()
	emunw.DriverInit()
	luabridge.DriverInit()
	retroarch.DriverInit()
	mock.DriverInit()

	// start the servers:
	grpcimpl.StartGrpcServer()
	usb2snes.StartHttpServer()

	// start up a systray:
	tray.CreateSystray()

	log.Println("main: exit")
}
