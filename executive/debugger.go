package executive

import (
	"fmt"
	"io"
	"net/http"

	"github.com/develatio/nebulant-cli/cast"
	ws "github.com/develatio/nebulant-cli/netproto/websocket"
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

func (d *debugger) Serve() error {
	// TODO: lookup for available port
	return http.ListenAndServe("localhost:6565", d)
}

func (d *debugger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		conn: conn,
		qq:   d.qq,
		vpty: nsterm.NewVirtPTY(),
		wsrw: ws.NewWebSocketReadWriteCloser(conn),
	}

	// c.clients = append(c.clients, clnt)
	go clnt.start()
}

// func (d *debugger) ExecCmd(cc *client) {
// 	// p := make([]byte, 250)
// 	// n, err := cc.Read(p)
// 	// if err != nil {
// 	// 	cast.LogErr(errors.Join(fmt.Errorf("err reading client data"), err).Error(), nil)
// 	// }
// 	// if n <= 0 {
// 	// 	return
// 	// }

// 	defer func() { go cc.Prompt() }()
// 	cmd := <-cc.cmdQueue
// 	cast.LogDebug(fmt.Sprintf("processing command %s", cmd), nil)
// 	fmt.Println(cmd)
// 	switch string(cmd) {
// 	case "shell":
// 		cc.stoppipe = make(chan struct{})
// 		defer close(cc.stoppipe)
// 		shell, err := term.DetermineOsShell()
// 		if err != nil {
// 			cast.LogErr(err.Error(), nil)
// 			cc.Write([]byte(err.Error()))
// 			return
// 		}
// 		cmd := exec.Command(shell)
// 		f, err := pty.Start(cmd)

// 		// f, err := os.OpenFile(filepath.Join("/dev", "tty"), unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0o620)
// 		// if err != nil {
// 		// 	cast.LogErr(err.Error(), nil)
// 		// 	cc.Write([]byte(err.Error()))
// 		// 	return
// 		// }

// 		// cmd.Env = append(os.Environ(), "TERM=vt100")
// 		// pipe, err := cmd.StdinPipe()
// 		if err != nil {
// 			cast.LogErr(err.Error(), nil)
// 			cc.Write([]byte(err.Error()))
// 			return
// 		}
// 		cc.stdinPipe = f
// 		// cmd.Stdin = f
// 		// stdin, err := cmd.StdinPipe()
// 		// if err != nil {
// 		// 	cast.LogErr(err.Error(), nil)
// 		// 	cc.Write([]byte(err.Error()))
// 		// }
// 		//		cc.stdinPipe = stdin

// 		// cmd.Stdout = cc
// 		// cmd.Stderr = cc
// 		// cmd.Stdin = os.Stdin
// 		// cmd.Stdout = os.Stdout
// 		// cmd.Stderr = os.Stderr

// 		cast.LogDebug(fmt.Sprintf("Running shell %v", shell), nil)
// 		go cc.PipeRead()
// 		io.Copy(cc, f)
// 		if err != nil {
// 			cast.LogErr(err.Error(), nil)
// 			cc.Write([]byte(err.Error()))
// 			return
// 		}
// 		cast.LogDebug(fmt.Sprintf("Stop of shell %v", shell), nil)
// 	default:
// 		cast.LogDebug("Unknown command "+string(cmd), nil)
// 		cc.Write([]byte("Unknown command " + string(cmd)))
// 	}

// 	// cast.LogDebug(fmt.Sprintf("New cmd from client: %s", p), nil)
// 	// cc.Write([]byte("Procesando...\n\r"))
// 	// time.Sleep(1 * time.Second)
// 	// cc.Write([]byte("resultado1\n\rresultado2\n\r"))
// 	// time.Sleep(1 * time.Second)
// }

type client struct {
	conn *websocket.Conn
	wsrw io.ReadWriteCloser
	qq   chan *client
	vpty *nsterm.VPTY2
}

func (c *client) start() {
	ldisc := nsterm.NewDefaultLdisc()
	c.vpty.SetLDisc(ldisc)

	// for wsock client
	// stdin and stdout for xterm.js
	mfd := c.vpty.MustarFD()
	go func() {
		io.Copy(mfd, c.wsrw)
		// fmt.Println("out of c to mfd")
	}()
	go func() {
		io.Copy(c.wsrw, mfd)
		// fmt.Println("out of mfd to c")
	}()

	// file descriptor for app
	sfd := c.vpty.SluvaFD()

	// welcome msg to debug session
	sfd.Write([]byte(term.Magenta + "Nebulant debug. Hello :)\r\nhow are you?\r\n" + term.Reset))

	// start shell
	// TODO: start debug program
	_, err := nsterm.NSShell(c.vpty, sfd, sfd)
	if err != nil {
		fmt.Println(err.Error())
	}
}

type breakPoint struct {
	end   chan bool
	stage *Stage
}
