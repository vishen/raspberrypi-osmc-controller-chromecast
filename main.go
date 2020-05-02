package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/vishen/go-chromecast/application"
)

// TODO: Add a default way to get all this in go-chromecast dns package.
type DNSEntry struct {
	UUID string
	Name string
	Addr string
	Port int
}

func (e DNSEntry) GetUUID() string {
	return e.UUID
}

func (e DNSEntry) GetName() string {
	return e.Name
}

func (e DNSEntry) GetAddr() string {
	return e.Addr
}

func (e DNSEntry) GetPort() int {
	return e.Port
}

func main() {

	// chromecast -- 192.168.86.30 8009
	// TODO: Get dynamically...
	e := DNSEntry{
		Addr: "192.168.86.30",
		Port: 8009,
	}

	iface, err := net.InterfaceByName("wlan0")
	if err != nil {
		log.Fatal(err)
	}

	app := application.NewApplication(iface, true, true)
	if err := app.Start(e); err != nil {
		log.Fatal(err)
	}

	filenames := []string{
		"/dev/hidraw0",
		"/dev/hidraw1",
	}

	resultsChan := make(chan *FunctionState, 1)
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		go func(r io.Reader) {
			var prevFunctionState *FunctionState
			for {
				functionState, err := parseFunctionState(r)
				if err != nil && err != io.EOF {
					log.Printf("error: unable to parse function from %q: %v", filename, err)
					continue
				}
				if functionState == nil {
					time.Sleep(2 * time.Second)
					continue
				}
				if prevFunctionState != nil && functionState.Action == Action_UP {
					functionState.Function = prevFunctionState.Function
				}
				resultsChan <- functionState
				prevFunctionState = functionState
			}
		}(f)
	}

	fmt.Printf("Waiting for input from controller...")
	fmt.Printf("app=%#v\n", app.Application())
	fmt.Printf("media=%#v\n", app.Media())
	fmt.Printf("volume=%#v\n", app.Volume())

	// TODO: Start a background goroutine that updates the state of
	// the chromecast every 0.5 seconds. Then holding the volume up
	// button will continually increase the volume. This can also
	// update the state of the internal chromecast representation
	// so it doesn't block the button being pushed.

	paused := false
	for fs := range resultsChan {
		app.Update()
		log.Printf("state=%#v\n", fs)
		fmt.Printf("app=%#v\n", app.Application())
		fmt.Printf("media=%#v\n", app.Media())
		fmt.Printf("volume=%#v\n", app.Volume())
		var err error
		switch fs.Function {
		case Function_STOP:
			if fs.Action == Action_DOWN {
				err = app.Stop()
			}
		case Function_PLAY_PAUSE:
			if fs.Action == Action_DOWN {
				if paused {
					err = app.Pause()
				} else {
					err = app.Unpause()
				}
				paused = !paused
			}
		case Function_ARROW_UP:
			if fs.Action == Action_DOWN {
				err = app.SetVolume(app.Volume().Level + 0.05)
			}
		case Function_ARROW_DOWN:
			if fs.Action == Action_DOWN {
				err = app.SetVolume(app.Volume().Level - 0.05)
			}
		case Function_ARROW_LEFT:
			if fs.Action == Action_DOWN {
				// TODO: ARROW_{LEFT,RIGHT} should be exponential if
				// pushed repeatedly over a short period of time.
				err = app.Seek(-1) // In Seconds
			}
		case Function_ARROW_RIGHT:
			if fs.Action == Action_DOWN {
				err = app.Seek(1) // In Seconds
			}
		case Function_OK:
			if fs.Action == Action_DOWN {
				err = app.SetMuted(!app.Volume().Muted)
			}
		case Function_PREV:
			if fs.Action == Action_DOWN {
				err = app.Previous()
			}
		case Function_NEXT:
			if fs.Action == Action_DOWN {
				err = app.Next()
			}
		}
		if err != nil {
			fmt.Printf("err: %#v\n", err)
		}
	}
}

type Action string

const (
	Action_DOWN Action = "DOWN"
	Action_UP   Action = "UP"
)

type Function string

const (
	Function_HOME        Function = "HOME"
	Function_ARROW_UP    Function = "ARROW_UP"
	Function_ARROW_LEFT  Function = "ARROW_LEFT"
	Function_ARROW_RIGHT Function = "ARROW_RIGHT"
	Function_ARROW_DOWN  Function = "ARROW_DOWN"
	Function_OK          Function = "OK"
	Function_INFO        Function = "INFO"
	Function_RETURN      Function = "RETURN"
	Function_PLAY_PAUSE  Function = "PLAY_PAUSE"
	Function_STOP        Function = "STOP"
	Function_PREV        Function = "PREV"
	Function_NEXT        Function = "NEXT"
	Function_LIST        Function = "LIST"
)

type FunctionState struct {
	Action    Action
	Function  Function
	Timestamp time.Time
}

func parseFunctionState(r io.Reader) (*FunctionState, error) {
	fs := &FunctionState{
		Action:    Action_DOWN,
		Timestamp: time.Now(),
	}
	b1 := make([]byte, 4)
	if _, err := r.Read(b1); err != nil {
		return fs, err
	}
	switch b1[0] {
	default:
		return nil, nil
	case 0x00:
		switch b1[2] {
		case 'J':
			fs.Function = Function_HOME
		case 'R':
			fs.Function = Function_ARROW_UP
		case 'P':
			fs.Function = Function_ARROW_LEFT
		case 'O':
			fs.Function = Function_ARROW_RIGHT
		case 'Q':
			fs.Function = Function_ARROW_DOWN
		case '(':
			fs.Function = Function_OK
		default:
			fs.Action = Action_UP
		}
	case 0x04:
		switch b1[1] {
		case '`':
			fs.Function = Function_INFO
		case '$':
			fs.Function = Function_RETURN
		case 0xcd:
			fs.Function = Function_PLAY_PAUSE
		case '&':
			fs.Function = Function_STOP
		case 0xb4:
			fs.Function = Function_PREV
		case 0xb3:
			fs.Function = Function_NEXT
		default:
			fs.Action = Action_UP
		}
	case 0x05:
		switch b1[1] {
		case 0x08:
			fs.Function = Function_LIST
		default:
			fs.Action = Action_UP
		}
	}
	return fs, nil
}
