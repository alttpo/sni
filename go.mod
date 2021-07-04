module sni

go 1.16

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/getlantern/systray v1.2.0
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.0.4
	github.com/spf13/viper v1.8.1
	go.bug.st/serial v1.1.3
	golang.org/x/sys v0.0.0-20210603125802-9665404d3644 // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/getlantern/systray => github.com/alttpo/systray v1.2.0
