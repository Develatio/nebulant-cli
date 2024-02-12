// Nebulant
// Copyright (C) 2024  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package runtime

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/creack/pty"
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
func NewDebugger(r *Runtime) *debugger {
	debugger := &debugger{
		// currentBreakPoint: breakpoint,
		// manager: breakpoint.stage.manager,
		//
		runtime: r,
		qq:      make(chan *client, 100),
		stop:    make(chan struct{}),
	}
	// debugger.breakPoints = append(debugger.breakPoints, breakpoint)
	Debuggers = append(Debuggers, debugger)
	return debugger
}

type debugger struct {
	runtime *Runtime
	// manager *Manager
	// breakPoints       []*breakPoint
	// currentBreakPoint *breakPoint
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
		dbg:  d,
		vpty: nsterm.NewVirtPTY(),
		wsrw: ws.NewWebSocketReadWriteCloser(conn),
	}

	// c.clients = append(c.clients, clnt)
	go clnt.start()
}

func (d *debugger) ExecCmd(cc *client, cmd string) {
	cast.LogDebug(fmt.Sprintf("processing command %s", cmd), nil)
	clientFD := cc.GetFD()

	switch string(cmd) {
	case "c":
		for bkpt := range d.runtime.GetBreakPoints() {
			bkpt.End()
		}
	case "ll":
		fmt.Fprintln(clientFD, "\tbreak points:")
		for bkpt := range d.runtime.GetBreakPoints() {
			actx := bkpt.GetActionContext()
			parents := actx.Parents()
			for _, parent := range parents {
				parents2 := parent.Parents()
				for _, parent := range parents2 {
					fmt.Fprintf(clientFD, "\t\t[%s]\n\t\t|\n", parent.GetAction().ActionName)
				}
				_act := parent.GetAction()
				if _act == nil {
					fmt.Fprintf(clientFD, "\t\t\t[null]\n\t\t\t|\n")
				}
				fmt.Fprintf(clientFD, "\t\t\t[%s]\n\t\t\t|\n", parent.GetAction().ActionName)
			}
			action := actx.GetAction()
			fmt.Fprintf(clientFD, "\t-> %p (id:%s)\n", bkpt, action.ActionName)
		}
	case "th":
		threads := d.runtime.GetThreads()
		for th := range threads {
			qq := th.GetQueue()
			fmt.Fprintf(clientFD, "\tthread %p\n", th)
			for _, actx := range qq {
				action := actx.GetAction()
				fmt.Fprintf(clientFD, "\t\t%s\n", action.ActionName)
			}
		}
	case "shell":
		// cc.stoppipe = make(chan struct{})
		// defer close(cc.stoppipe)

		shell, err := term.DetermineOsShell()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			fmt.Fprintln(clientFD, err.Error())
			return
		}
		cmd := exec.Command(shell)
		f, err := pty.Start(cmd)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			fmt.Fprintln(clientFD, err.Error())
			return
		}

		fmt.Fprintln(clientFD, "initializing local shell...")

		cc.Raw(true)
		defer cc.Raw(false)

		cast.LogDebug(fmt.Sprintf("Running shell %v", shell), nil)

		go func() { _, _ = io.Copy(f, clientFD) }()
		_, err = io.Copy(clientFD, f)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			fmt.Fprintln(clientFD, err.Error())
			return
		}
		cast.LogDebug(fmt.Sprintf("Stop of shell %v", shell), nil)
	default:
		cast.LogDebug("Unknown command "+string(cmd), nil)
		fmt.Fprintln(clientFD, "Unknown command "+string(cmd))
	}
}

type client struct {
	dbg  *debugger
	conn *websocket.Conn
	wsrw io.ReadWriteCloser
	vpty *nsterm.VPTY2
	// default ldisc, used in Raw(false)
	ldisc   nsterm.Ldisc
	sluvaFD *nsterm.PortFD
}

// func (c *client) Write(p []byte) (n int, err error) {
// 	return c.stdout.Write(p)
// }

func (c *client) Raw(activate bool) {
	if activate {
		c.vpty.SetLDisc(nsterm.NewRawLdisc())
		return
	}
	c.vpty.SetLDisc(c.ldisc)
}

func (c *client) GetFD() io.ReadWriteCloser {
	return c.sluvaFD
}

func (c *client) start() {
	ldisc := nsterm.NewDefaultLdisc()
	c.vpty.SetLDisc(ldisc)
	c.ldisc = ldisc

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
	c.sluvaFD = sfd

	// welcome msg to debug session
	sfd.Write([]byte(term.Magenta + "Nebulant debug. Hello :)\r\nhow are you?\r\n" + term.Reset))

	prmpt := nsterm.NewPrompt(c.vpty, sfd, sfd)
	prmpt.SetPS1("Nebulant dbg> ")
	for {
		s, err := prmpt.ReadLine()
		if err != nil {
			// TODO: disconnect with err
		}
		if s == nil {
			// no command
			continue
		}

		// built in :)
		if *s == "exit" {
			// TODO: disconnect
		}

		// built in ;)
		if *s == "help" {
			// show debug help
		}

		c.dbg.ExecCmd(c, *s)
	}
}
