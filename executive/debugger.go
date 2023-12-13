package executive

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
	"unicode/utf8"

	"github.com/creack/pty"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/nsterm"
	"github.com/develatio/nebulant-cli/term"
	"github.com/gorilla/websocket"
)

var Debuggers []*debugger

var MAXWRITESIZE = 1024
var MAXREADSIZE = 1024

// Newdebugger func
func NewDebugger(breakpoint *breakPoint) *debugger {
	debugger := &debugger{
		currentBreakPoint: breakpoint,
		manager:           breakpoint.stage.manager,
		//
		qq:   make(chan *client, 100),
		stop: make(chan struct{}),
		// execInstruction:   make(chan *ExecCtrlInstruction, 10),
	}
	debugger.breakPoints = append(debugger.breakPoints, breakpoint)
	Debuggers = append(Debuggers, debugger)
	return debugger
}

type debugger struct {
	manager *Manager
	// execInstruction   chan *ExecCtrlInstruction
	breakPoints       []*breakPoint
	currentBreakPoint *breakPoint
	//
	// stdin  io.Reader
	// stdout io.Writer
	// stderr io.Writer
	//
	qq   chan *client
	stop chan struct{}
}

func (c *debugger) Serve() error {
	// lookup for available port
	go c.Scan()
	return http.ListenAndServe("localhost:6565", c)
}

func (c *debugger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cast.LogInfo("Debug client in", nil)
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  MAXREADSIZE,
		WriteBufferSize: MAXWRITESIZE,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		cast.LogErr(err.Error(), nil)
		return
	}

	clnt := &client{
		cmdQueue: make(chan []byte, 525),
		conn:     conn,
		qq:       c.qq,
		vpty:     nsterm.NewVirtPTY(),
	}

	// for wsock client
	// stdin and stdout for xterm.js
	// mfd := clnt.vpty.MustarFD()
	// // file descriptor for app
	// sfd := clnt.vpty.SluvaFD()

	// c.clients = append(c.clients, clnt)
	clnt.Write([]byte(term.Magenta + "Nebulant debug. Hello :)\r\nhow are you?\r\n" + term.Reset))
	go clnt.start()

}

func (c *debugger) Scan() {
L:
	for { // Infine loop until break L
		select { // Loop until a case ocurrs.
		case cc := <-c.qq:
			go c.ExecCmd(cc)
		case <-c.stop:
			// on close(c.stop)
			break L
		default:
			time.Sleep(200 * time.Millisecond)

		}
	}
}

func (c *debugger) ExecCmd(cc *client) {
	// p := make([]byte, 250)
	// n, err := cc.Read(p)
	// if err != nil {
	// 	cast.LogErr(errors.Join(fmt.Errorf("err reading client data"), err).Error(), nil)
	// }
	// if n <= 0 {
	// 	return
	// }

	defer func() { go cc.Prompt() }()
	cmd := <-cc.cmdQueue
	cast.LogDebug(fmt.Sprintf("processing command %s", cmd), nil)
	fmt.Println(cmd)
	switch string(cmd) {
	case "shell":
		cc.stoppipe = make(chan struct{})
		defer close(cc.stoppipe)
		shell, err := term.DetermineOsShell()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cc.Write([]byte(err.Error()))
			return
		}
		cmd := exec.Command(shell)
		f, err := pty.Start(cmd)

		// f, err := os.OpenFile(filepath.Join("/dev", "tty"), unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0o620)
		// if err != nil {
		// 	cast.LogErr(err.Error(), nil)
		// 	cc.Write([]byte(err.Error()))
		// 	return
		// }

		// cmd.Env = append(os.Environ(), "TERM=vt100")
		// pipe, err := cmd.StdinPipe()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cc.Write([]byte(err.Error()))
			return
		}
		cc.stdinPipe = f
		// cmd.Stdin = f
		// stdin, err := cmd.StdinPipe()
		// if err != nil {
		// 	cast.LogErr(err.Error(), nil)
		// 	cc.Write([]byte(err.Error()))
		// }
		//		cc.stdinPipe = stdin

		// cmd.Stdout = cc
		// cmd.Stderr = cc
		// cmd.Stdin = os.Stdin
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr

		cast.LogDebug(fmt.Sprintf("Running shell %v", shell), nil)
		go cc.PipeRead()
		io.Copy(cc, f)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cc.Write([]byte(err.Error()))
			return
		}
		cast.LogDebug(fmt.Sprintf("Stop of shell %v", shell), nil)
	default:
		cast.LogDebug("Unknown command "+string(cmd), nil)
		cc.Write([]byte("Unknown command " + string(cmd)))
	}

	// cast.LogDebug(fmt.Sprintf("New cmd from client: %s", p), nil)
	// cc.Write([]byte("Procesando...\n\r"))
	// time.Sleep(1 * time.Second)
	// cc.Write([]byte("resultado1\n\rresultado2\n\r"))
	// time.Sleep(1 * time.Second)
}

// func (c *debuggerMultiStdDispatch) Read(p []byte) (n int, err error) {}

// func (c *debuggerMultiStdDispatch) Read(p []byte) (n int, err error) {
// 	cast.LogDebug("debugger: Reading from clients", nil)
// 	for {
// 		for _, clnt := range c.clients {
// 			ppp := make([]byte, 255)
// 			n, err := clnt.Read(ppp)
// 			if err != nil {
// 				cast.LogErr(errors.Join(fmt.Errorf("err reading from debug client"), err).Error(), nil)
// 				continue
// 			}
// 			if n <= 0 {
// 				continue
// 			} else {
// 				fmt.Println(ppp)
// 				// eco to all clients
// 				cast.LogDebug(fmt.Sprintf("echoing %s", ppp), nil)
// 				_, err := c.Write(ppp)
// 				if err != nil {
// 					cast.LogErr(errors.Join(fmt.Errorf("err writin eco debug client"), err).Error(), nil)
// 					continue
// 				}
// 				cast.LogDebug(fmt.Sprintf("reading %v bytes from client", n), nil)
// 				p = ppp
// 				return n, nil
// 			}

// 		}
// 		time.Sleep(500 * time.Millisecond)
// 	}

// 	return 0, nil
// }

type client struct {
	cmdQueue chan []byte
	//
	stdinPipe io.WriteCloser
	// stdinReader *io.PipeReader
	//
	conn       *websocket.Conn
	qq         chan *client
	stopprompt chan struct{}
	// pipetick   chan struct{}
	stoppipe chan struct{}
	// w io.Writer
	// r io.Reader
	vpty *nsterm.VPTY2
}

func (c *client) start() {
	ldisc := nsterm.NewDefaultLdisc()
	c.vpty.SetLDisc(ldisc)

	// for wsock client
	// stdin and stdout for xterm.js
	mfd := c.vpty.MustarFD()
	go func() {
		io.Copy(mfd, c)
		fmt.Println("out of c to mfd")
	}()
	go func() {
		io.Copy(c, mfd)
		fmt.Println("out of mfd to c")
	}()

	// file descriptor for app
	sfd := c.vpty.SluvaFD()
	_, err := nsterm.NSShell(c.vpty, sfd, sfd)
	if err != nil {
		fmt.Println(err.Error())
	}

	// PS1 := []byte("(DBG) NBShell> ")

	// for {
	// 	stdout.Write([]byte(term.CursorToColZero + term.EraseEntireLine))
	// 	stdout.Write(PS1)
	// 	stdout.Write([]byte(string(ldisc.RuneBuff)))
	// L2:
	// 	for {
	// 		esc := <-ldisc.ESC
	// 		if esc != nsterm.CarriageReturn {
	// 			continue
	// 		}

	// 	}

}

func (c *client) Write(p []byte) (int, error) {
	buff := bytes.NewBuffer(p)
	n := 0
	// write in chunks
	for {
		p := make([]byte, MAXWRITESIZE)
		_n, err := buff.Read(p)
		n = n + _n
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
		err = c.conn.WriteMessage(websocket.BinaryMessage, p)
		if err != nil {
			return n, err
		}
	}
}

func (c *client) Read(p []byte) (n int, err error) {
	_, m, err := c.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	ml := len(m)
	pl := len(p)
	fl := ml
	if ml > pl {
		fl = pl
	}
	for i := 0; i < fl; i++ {
		p[i] = m[i]
	}

	return len(p), nil
}

func (c *client) PipeRead() error {
	// c.pipetick = make(chan struct{}, 10)
L:
	for {
		cast.LogDebug("Reading from pipe", nil)
		select { // Loop until a case ocurrs.
		case <-c.stoppipe:
			cast.LogDebug("Out of pipe read", nil)
			break L
		default:
			cast.LogDebug("Reading from ws...", nil)
			mt, m, err := c.conn.ReadMessage()
			if err != nil {
				cast.LogErr(err.Error(), nil)
			}
			cast.LogDebug(fmt.Sprintf("pipe msg type %d", mt), nil)

			if mt != websocket.TextMessage {
				// handle non binary messages
				continue
			}

			mr, _ := utf8.DecodeRune(m)
			cast.LogDebug(fmt.Sprintf("rune %U", mr), nil)

			// m = bytes.Replace(m, []byte{13, 10}, []byte{13}, -1)
			// m = bytes.Replace(m, []byte{13}, []byte{13, 10}, -1)

			fmt.Println(mr, m, []byte("\n"))

			if mr == 13 {
				mr = 10
			}
			cast.LogDebug(fmt.Sprintf("write to stdinwriter: %s", m), nil)
			_, err = c.stdinPipe.Write([]byte(string(mr)))
			if err != nil {
				cast.LogErr(err.Error(), nil)
			}
			cast.LogDebug(fmt.Sprintf("writed %s", m), nil)
			// c.pipeBuff.Write(m)
			// eco
			// _, err = c.Write(m)
			// if err != nil {
			// 	cast.LogErr(err.Error(), nil)
			// }
		}
	}
	return nil
}

func (c *client) Prompt() error {
	c.stopprompt = make(chan struct{})
	c.Write([]byte("Nebulant debug > "))
	var bff []byte
	buff := bytes.NewBuffer(bff)
L:
	for { // Infine loop until break L
		select { // Loop until a case ocurrs.
		case <-c.stopprompt:
			break L
		default:
			cast.LogDebug("loop of prompt...", nil)
			mt, m, err := c.conn.ReadMessage()
			if err != nil {
				return err
			}
			if mt != websocket.TextMessage {
				cast.LogDebug(fmt.Sprintf("non binary msg type %v", mt), nil)
				// handle non binary messages
				continue
			}

			mr, _ := utf8.DecodeRune(m)
			cast.LogDebug(fmt.Sprintf("rune %U", mr), nil)

			// CRLF -> CR -> CRLF
			// m := bytes.Replace(message, []byte{13, 10}, []byte{13}, -1)
			// m = bytes.Replace(m, []byte{13}, []byte{13, 10}, -1)

			if mr == 13 {
				mr = 10
			}

			// ^C U+000D
			if mr == 3 {
				c.conn.Close()
				break L
			}

			// del U+007F
			if mr == 127 {
				if buff.Len() <= 0 {
					continue
				}
				// eco
				_, err = c.Write([]byte("\b" + term.EraseLineFromCursor))
				if err != nil {
					return err
				}
				buff.Truncate(buff.Len() - 1)
				cast.LogDebug(fmt.Sprintf("New buff: %s", buff.Bytes()), nil)
				continue
			}

			// debug
			fmt.Println(m)
			// eco //
			_, err = c.Write(m)
			if err != nil {
				return err
			}
			// // //

			// process
			if mr == 10 {
				cast.LogDebug("processing command", nil)
				c.cmdQueue <- buff.Bytes()
				c.qq <- c
				break L
			}

			_, err = buff.WriteRune(mr)
			if err != nil {
				return err
			}

			// test for carriage return
			// cut by "\n"
			// send all parts to c.cmdQueue except
			// last part if it dsnt have \n
			// keep storing client data
			//
			//
			//
		}
	}
	return nil
}

type breakPoint struct {
	end   chan bool
	stage *Stage
	//
}
