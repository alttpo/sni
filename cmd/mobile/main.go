package mobile

import (
	"log"
	"sni/cmd/sni/appversion"
	"sni/cmd/sni/config"
	"sni/devices/snes/drivers/emunwa"
	"sni/devices/snes/drivers/fxpakpro"
	"sni/devices/snes/drivers/luabridge"
	"sni/devices/snes/drivers/mock"
	"sni/devices/snes/drivers/retroarch"
	"sni/services/grpcimpl"
	"sni/services/usb2snes"
	"sync"
)

// build variables set via ldflags by `go build -ldflags="-X 'main.version=v1.0.0'"`:
var (
	version string = "v0.0.0"
	commit  string = "dirty"
	date    string = "2021-05-03T00:17:00Z"
	builtBy string = "go"
)

var once sync.Once

func Start() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recover from Start(): %v\n", err)
		}
	}()

	once.Do(func() {
		// make the version info public in the appversion package because the main package cannot be imported:
		appversion.Init(
			version,
			commit,
			date,
			builtBy,
		)

		// TODO: initialize logging subsystem:
		//logging.Init()
		//log.SetOutput(io.Discard)

		// TODO: load configuration:
		//config.Load()
		config.VerboseLogging = true

		// explicitly initialize all the drivers:
		fxpakpro.DriverInit()
		emunwa.DriverInit()
		luabridge.DriverInit()
		retroarch.DriverInit()
		mock.DriverInit()
	})

	// start the servers:
	grpcimpl.StartGrpcServer()
	usb2snes.StartHttpServer()
}

func Stop() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recover from Stop(): %v\n", err)
		}
	}()

	grpcimpl.StopGrpcServer()
	// TODO: Close() on usb2snes's `*http.Server`s
}
