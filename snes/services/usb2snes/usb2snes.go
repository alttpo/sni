package usb2snes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sni/cmd/sni/appversion"
	"sni/cmd/sni/config"
	"sni/cmd/sni/tray"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"sni/util"
	"sni/util/env"
	"strconv"
	"strings"
	"time"
)

func StartHttpServer() {
	// Parse env vars:
	disabled := env.GetOrDefault("SNI_USB2SNES_DISABLE", "0")
	if util.IsTruthy(disabled) {
		log.Printf("usb2snes: server disabled due to env var %s=%s\n", "SNI_USB2SNES_DISABLE", disabled)
		return
	}

	addrList := env.GetOrDefault("SNI_USB2SNES_LISTEN_ADDRS", "0.0.0.0:23074,0.0.0.0:8080")
	listenAddrs := strings.Split(addrList, ",")
	for _, listenAddr := range listenAddrs {
		go listenHttp(listenAddr)
	}
}

func listenHttp(listenAddr string) {
	var err error
	var lis net.Listener

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(clientWebsocketHandler))

	// attempt to start the usb2snes server:
	count := 0
	lc := &net.ListenConfig{Control: util.ReusePortControl}
	for {
		lis, err = lc.Listen(context.Background(), "tcp", listenAddr)
		if err == nil {
			break
		}

		if count == 0 {
			log.Printf("usb2snes: failed to listen on %s: %v\n", listenAddr, err)
		}
		count++
		if count >= 30 {
			count = 0
		}

		time.Sleep(time.Second)
	}

	log.Printf("usb2snes: listening on %s\n", listenAddr)
	err = http.Serve(lis, mux)
	log.Printf("usb2snes: exit listenHttp: %v\n", err)
}

type wsReader struct {
	r *wsutil.Reader
}

func (w *wsReader) Read(p []byte) (n int, err error) {
	n, err = w.r.Read(p)
	// if we need to advance frame to continue reading, do so:
	if err == wsutil.ErrNoFrameAdvance {
		var hdr ws.Header
		hdr, err = w.r.NextFrame()
		if err != nil {
			return
		}
		if hdr.OpCode == ws.OpClose {
			return n, io.EOF
		}

		// we must read here to fulfill the original request:
		if n == 0 {
			n, err = w.r.Read(p)
		}
		return
	}

	// clear EOF error when reading to the end of the frame:
	if err == io.EOF && n > 0 {
		err = nil
	}

	return
}

type wsWriter struct {
	w         *wsutil.Writer
	frameSize int
	written   int
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.written += n
	if err != nil {
		return
	}

	if w.written >= w.frameSize {
		err = w.w.Flush()
		w.written = 0
	}
	return
}

func clientWebsocketHandler(rw http.ResponseWriter, req *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(req, rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(400)
		return
	}

	clientName := conn.RemoteAddr().String()
	defer func() {
		log.Printf("usb2snes: %s: %s disconnected\n", clientName, conn.RemoteAddr())
		conn.Close()
	}()

	// setup general readers, writers and JSON encoders, decoders:
	r := wsutil.NewReader(conn, ws.StateServerSide)
	wb := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpBinary)
	wj := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
	je := json.NewEncoder(wj)
	jd := json.NewDecoder(r)

	var attachedUri *url.URL
	var driver snes.Driver
	var device snes.AutoCloseableDevice
	var deviceMemoryMapping sni.MemoryMapping

	_ = driver

	log.Printf("usb2snes: %s: connected\n", conn.RemoteAddr())

serverLoop:
	for {
		hdr, err := r.NextFrame()
		if err == io.EOF {
			log.Printf("usb2snes: %s: client closed connection\n", clientName)
			break serverLoop
		}
		if err != nil {
			log.Printf("usb2snes: %s: error reading next websocket frame: %s\n", clientName, err)
			break serverLoop
		}
		if hdr.OpCode == ws.OpClose {
			log.Printf("usb2snes: %s: client closed connection\n", clientName)
			break serverLoop
		}
		if hdr.OpCode == ws.OpPing {
			var p []byte
			p, err = io.ReadAll(r)
			if err != nil {
				log.Printf("usb2snes: %s: error reading ping frame: %s\n", clientName, err)
				break serverLoop
			}
			err = ws.WriteFrame(conn, ws.NewPongFrame(p))
			if err != nil {
				log.Printf("usb2snes: %s: error writing pong frame: %s\n", clientName, err)
				break serverLoop
			}
			continue serverLoop
		}

		if hdr.OpCode != ws.OpText {
			log.Printf("usb2snes: %s: client sent unexpected websocket frame opcode 0x%x\n", clientName, hdr.OpCode)
			err = r.Discard()
			if err != nil {
				log.Printf("usb2snes: %s: error discarding websocket frame: %s\n", clientName, err)
				break serverLoop
			}
			continue serverLoop
		}

		type command struct {
			Opcode   string   `json:"Opcode"`
			Space    string   `json:"Space"`
			Flags    []string `json:"Flags,omitempty"`
			Operands []string `json:"Operands,omitempty"`
		}
		var cmd command
		err = jd.Decode(&cmd)
		if err != nil {
			log.Printf("usb2snes: %s: could not decode json request: %s\n", clientName, err)
			break serverLoop
		}

		type response struct {
			Results []string `json:"Results"`
		}
		var results response

		if config.VerboseLogging {
			log.Printf("usb2snes: %s: %s %s [%s]\n", clientName, cmd.Opcode, cmd.Space, strings.Join(cmd.Operands, ","))
		}

		replyJson := func() bool {
			if config.VerboseLogging {
				log.Printf("usb2snes: %s: %s REPLY: %+v\n", clientName, cmd.Opcode, results)
			}

			err = je.Encode(results)
			if err != nil {
				log.Printf("usb2snes: %s: %s error encoding json response: %s\n", clientName, cmd.Opcode, err)
				return false
			}
			if err = wj.Flush(); err != nil {
				log.Printf("usb2snes: %s: %s error flushing response: %s\n", clientName, cmd.Opcode, err)
				return false
			}
			return true
		}

		switch cmd.Opcode {
		case "DeviceList":
			descriptors := make([]snes.DeviceDescriptor, 0, 10)
			for _, driver := range snes.Drivers() {
				if config.VerboseLogging {
					log.Printf("usb2snes: %s: %s detecting devices from driver '%s'\n", clientName, cmd.Opcode, driver.Name)
				}
				d, err := driver.Driver.Detect()
				if err != nil {
					log.Printf("usb2snes: %s: %s error detecting from driver '%s': %s\n", clientName, cmd.Opcode, driver.Name, err)
					continue
				}
				descriptors = append(descriptors, d...)
			}
			if config.VerboseLogging {
				log.Printf("usb2snes: %s: %s detection complete\n", clientName, cmd.Opcode)
			}

			go tray.UpdateDeviceList(descriptors)

			results.Results = make([]string, 0, 10)
			for _, descriptor := range descriptors {
				results.Results = append(results.Results, descriptor.Uri.String())
			}

			if !replyJson() {
				break serverLoop
			}
			break
		case "Name":
			if len(cmd.Operands) != 1 {
				log.Printf("usb2snes: %s: %s missing required operand\n", clientName, cmd.Opcode)
				break serverLoop
			}

			clientName = cmd.Operands[0]
			log.Printf("usb2snes: %s: %s '%s'\n", conn.RemoteAddr(), cmd.Opcode, clientName)
			break
		case "AppVersion":
			results.Results = []string{fmt.Sprintf("SNI-%s", appversion.Version)}

			if !replyJson() {
				break serverLoop
			}
			break
		case "Close":
			break serverLoop

		case "Attach":
			if len(cmd.Operands) != 1 {
				log.Printf("usb2snes: %s: missing required Operand\n", clientName)
				break serverLoop
			}

			uriString := strings.TrimSpace(cmd.Operands[0])
			attachedUri, err = url.Parse(uriString)
			if err != nil {
				log.Printf("usb2snes: %s: bad device uri '%s': %s\n", clientName, uriString, err)
				break serverLoop
			}

			driver, device, err = snes.DeviceByUri(attachedUri)
			if err != nil {
				log.Printf("usb2snes: %s: could not open device by uri '%s': %s\n", clientName, uriString, err)
				break serverLoop
			}

			//var confidence bool
			//var outHeaderBytes []byte
			deviceMemoryMapping, _, _, err = mapping.Detect(context.Background(), device, nil, nil)
			if err != nil {
				log.Printf("usb2snes: %s: could not detect memory mapping: %s\n", clientName, err)
				break serverLoop
			}
			break
		case "Info":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", conn.RemoteAddr(), cmd.Opcode)
				break serverLoop
			}

			// TODO:
			results.Results = []string{"1.9.0-usb-v9", "SD2SNES", "No Info"}

			if !replyJson() {
				break serverLoop
			}
			break
		case "GetAddress":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 2 {
				log.Printf("usb2snes: %s: %s expected at least 2 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}
			if len(cmd.Operands)&1 != 0 {
				log.Printf("usb2snes: %s: %s expected even number of operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			// parse operands as (addr, size) pairs:
			ops := cmd.Operands[:]
			reqs := make([]snes.MemoryReadRequest, len(ops)/2)
			for i := 0; i < len(reqs); i++ {
				addrHex := cmd.Operands[i*2]
				var addr uint64
				addr, err = strconv.ParseUint(addrHex, 16, 32)
				if err != nil {
					log.Printf("usb2snes: %s: %s: bad operand [%d]: '%s'\n", clientName, cmd.Opcode, i*2, addrHex)
					break serverLoop
				}

				sizeHex := cmd.Operands[i*2+1]
				var size uint64
				size, err = strconv.ParseUint(sizeHex, 16, 32)
				if err != nil {
					log.Printf("usb2snes: %s: %s: bad operand [%d]: '%s'\n", clientName, cmd.Opcode, i*2+1, sizeHex)
					break serverLoop
				}

				var addr32 uint32
				space := strings.TrimSpace(strings.ToUpper(cmd.Space))
				switch space {
				case "SNES":
					addr32 = uint32(addr & 0x00_FFFFFF)
					break
				case "CMD":
					// dirty dirty hack to put the CMD address space into the FxPakPro space as some sort of subspace:
					addr32 = uint32(addr&0x00_FFFFFF) | 0x01_000000
					break
				default:
					log.Printf("usb2snes: %s: %s: unrecognized space '%s'\n", clientName, cmd.Opcode, space)
					break serverLoop
				}

				reqs[i] = snes.MemoryReadRequest{
					RequestAddress: snes.AddressTuple{
						Address:       addr32,
						AddressSpace:  sni.AddressSpace_FxPakPro,
						MemoryMapping: deviceMemoryMapping,
					},
					Size: int(size),
				}
			}

			// issue the read request:
			var rsps []snes.MemoryReadResponse
			rsps, err = device.MultiReadMemory(context.Background(), reqs...)
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}

			// write the response data:
			for i := range rsps {
				_, err = wb.Write(rsps[i].Data)
				if err != nil {
					log.Printf("usb2snes: %s: %s error writing response data: %s\n", clientName, cmd.Opcode, err)
					break serverLoop
				}
			}
			if config.VerboseLogging {
				log.Printf("usb2snes: %s: %s REPLY\n", clientName, cmd.Opcode)
			}

			if err = wb.Flush(); err != nil {
				log.Printf("usb2snes: %s: %s error flushing response: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break
		case "PutAddress":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 2 {
				log.Printf("usb2snes: %s: %s expected at least 2 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}
			if len(cmd.Operands)&1 != 0 {
				log.Printf("usb2snes: %s: %s expected even number of operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			// parse operands as (addr, size) pairs:
			ops := cmd.Operands[:]
			reqs := make([]snes.MemoryWriteRequest, len(ops)/2)
			reqCount := len(reqs)
			for i := 0; i < reqCount; i++ {
				addrHex := cmd.Operands[i*2]
				var addr uint64
				addr, err = strconv.ParseUint(addrHex, 16, 32)
				if err != nil {
					log.Printf("usb2snes: %s: %s: bad operand [%d]: '%s'\n", clientName, cmd.Opcode, i*2, addrHex)
					break serverLoop
				}

				sizeHex := cmd.Operands[i*2+1]
				var size uint64
				size, err = strconv.ParseUint(sizeHex, 16, 32)
				if err != nil {
					log.Printf("usb2snes: %s: %s: bad operand [%d]: '%s'\n", clientName, cmd.Opcode, i*2+1, sizeHex)
					break serverLoop
				}

				var addr32 uint32
				space := strings.TrimSpace(strings.ToUpper(cmd.Space))
				switch space {
				case "SNES":
					addr32 = uint32(addr & 0x00_FFFFFF)
					break
				case "CMD":
					// dirty dirty hack to put the CMD address space into the FxPakPro space as some sort of subspace:
					addr32 = uint32(addr&0x00_FFFFFF) | 0x01_000000
					break
				default:
					log.Printf("usb2snes: %s: %s: unrecognized space '%s'\n", clientName, cmd.Opcode, space)
					break serverLoop
				}

				reqs[i] = snes.MemoryWriteRequest{
					RequestAddress: snes.AddressTuple{
						Address:       addr32,
						AddressSpace:  sni.AddressSpace_FxPakPro,
						MemoryMapping: deviceMemoryMapping,
					},
					Data: make([]byte, size),
				}

				var n int
				wsr := &wsReader{r: r}
				n, err = wsr.Read(reqs[i].Data)
				_ = n
				//log.Printf("usb2snes: %s: %s read()[%d/%d]: read %d bytes; expected %d\n", clientName, cmd.Opcode, i+1, reqCount, n, size)
				if err != nil && err != io.EOF {
					log.Printf("usb2snes: %s: %s read()[%d/%d]: %s\n", clientName, cmd.Opcode, i+1, reqCount, err)
					break serverLoop
				}
			}

			// issue the read request:
			var rsps []snes.MemoryWriteResponse
			rsps, err = device.MultiWriteMemory(context.Background(), reqs...)
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			if config.VerboseLogging {
				log.Printf("usb2snes: %s: %s REPLY: %+v\n", clientName, cmd.Opcode, rsps)
			}

			_ = rsps
			break

		case "Reset":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			err = device.ResetSystem(context.Background())
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		case "Boot":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 1 {
				log.Printf("usb2snes: %s: %s expected 1 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			err = device.BootFile(context.Background(), cmd.Operands[0])
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		case "List":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 1 {
				log.Printf("usb2snes: %s: %s expected 1 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			var entries []snes.DirEntry
			entries, err = device.ReadDirectory(context.Background(), cmd.Operands[0])
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}

			// translate entries into string array:
			for _, entry := range entries {
				results.Results = append(results.Results, strconv.Itoa(int(entry.Type)), entry.Name)
			}

			if !replyJson() {
				break serverLoop
			}
			break

		case "MakeDir":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 1 {
				log.Printf("usb2snes: %s: %s expected 1 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			err = device.MakeDirectory(context.Background(), cmd.Operands[0])
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		case "Remove":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 1 {
				log.Printf("usb2snes: %s: %s expected 1 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			err = device.RemoveFile(context.Background(), cmd.Operands[0])
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		case "Rename":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 2 {
				log.Printf("usb2snes: %s: %s expected 2 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			err = device.RenameFile(context.Background(), cmd.Operands[0], cmd.Operands[1])
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		case "GetFile":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 1 {
				log.Printf("usb2snes: %s: %s expected 1 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			var progress snes.ProgressReportFunc = nil
			if config.VerboseLogging {
				progress = func(current uint32, total uint32) {
					log.Printf("usb2snes: %s: %s: progress $%08x/$%08x\n", clientName, cmd.Opcode, current, total)
				}
			}

			// calls Flush after every `frameSize` bytes written:
			wsw := &wsWriter{w: wb, frameSize: 1024}

			// callback to write the json reply about the file size:
			sizeReceived := snes.SizeReceivedFunc(func(size uint32) {
				// send a single response packet with the size of the file
				results.Results = append(results.Results, strconv.FormatUint(uint64(size), 16))
				if !replyJson() {
					// TODO: cause failure of writer
				}
			})

			var n uint32
			n, err = device.GetFile(context.Background(), cmd.Operands[0], wsw, sizeReceived, progress)
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			if config.VerboseLogging {
				log.Printf("usb2snes: %s: %s REPLY: $%x bytes\n", clientName, cmd.Opcode, n)
			}
			if err = wb.Flush(); err != nil {
				log.Printf("usb2snes: %s: %s error flushing response: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		case "PutFile":
			if device == nil {
				log.Printf("usb2snes: %s: %s requires Attach first\n", clientName, cmd.Opcode)
				break serverLoop
			}

			if len(cmd.Operands) < 2 {
				log.Printf("usb2snes: %s: %s expected 2 operands, got %d\n", clientName, cmd.Opcode, len(cmd.Operands))
				break serverLoop
			}

			var size64 uint64
			size64, err = strconv.ParseUint(cmd.Operands[1], 16, 32)
			if err != nil {
				log.Printf("usb2snes: %s: %s: bad operand [%d]: '%s'\n", clientName, cmd.Opcode, 1, cmd.Operands[1])
				break serverLoop
			}
			size := uint32(size64)

			var progress snes.ProgressReportFunc = nil
			if config.VerboseLogging {
				progress = func(current uint32, total uint32) {
					log.Printf("usb2snes: %s: %s: progress $%08x/$%08x\n", clientName, cmd.Opcode, current, total)
				}
			}

			var n uint32
			wsr := &wsReader{r}
			n, err = device.PutFile(context.Background(), cmd.Operands[0], size, wsr, progress)
			if err != nil {
				log.Printf("usb2snes: %s: %s error: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			if config.VerboseLogging {
				log.Printf("usb2snes: %s: %s REPLY: $%x bytes\n", clientName, cmd.Opcode, n)
			}
			if err = wb.Flush(); err != nil {
				log.Printf("usb2snes: %s: %s error flushing response: %s\n", clientName, cmd.Opcode, err)
				break serverLoop
			}
			break

		default:
			log.Printf("usb2snes: %s: unrecognized opcode '%s'\n", clientName, cmd.Opcode)
			break
		}
	}
}
