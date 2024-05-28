// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package runtime

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	ws "github.com/develatio/nebulant-cli/netproto/websocket"
	"github.com/develatio/nebulant-cli/nhttpd"
	"github.com/develatio/nebulant-cli/nsterm"
	"github.com/develatio/nebulant-cli/storage"
	nebulant_term "github.com/develatio/nebulant-cli/term"
	"github.com/google/shlex"
	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

var debuggers map[*Runtime]*debugger = make(map[*Runtime]*debugger)

var MAXWRITESIZE = 1024
var MAXREADSIZE = 1024

// TODO: merge with interactive.browser
// serializers and move to common path
type newBridgePoolSerializer struct {
	Token        string `json:"token"`
	Scheme       string `json:"scheme"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ConsumerPath string `json:"consumer_path"`
	XtermPath    string `json:"xterm_path"`
	CliPath      string `json:"cli_path"`
	cliURL       *url.URL
}

// like user-space
type clientSpace struct {
	store base.IStore
}

// Newdebugger func
func NewDebugger(r *Runtime) *debugger {
	if _, exists := debuggers[r]; exists {
		return debuggers[r]
	}
	debugger := &debugger{
		// currentBreakPoint: breakpoint,
		// manager: breakpoint.stage.manager,
		//
		runtime:   r,
		qq:        make(chan *client, 100),
		stop:      make(chan struct{}),
		rclients:  make(map[*client]bool),
		elistener: r.NewEventListener(),
	}
	// debugger.breakPoints = append(debugger.breakPoints, breakpoint)
	debuggers[r] = debugger
	return debugger
}

func (d *debugger) SetCursor(actx base.IActionContext) {
	d.cursor = actx
}

func (d *debugger) GetCursor() base.IActionContext {
	return d.cursor
}

type debugger struct {
	// used to block debugger
	waiting bool
	runtime *Runtime
	// manager *Manager
	// breakPoints       []*breakPoint
	// currentBreakPoint *breakPoint
	detach    []*nsterm.VPTY2
	running   bool
	remote    bool
	rclients  map[*client]bool
	lclient   *client
	cursor    base.IActionContext
	qq        chan *client
	stop      chan struct{}
	close     chan error
	elistener *base.EventListener
}

func (d *debugger) l(msg string) {
	cast.LogInfo(fmt.Sprintf("DBG: %s", msg), d.runtime.irb.ExecutionUUID)
}

func (d *debugger) le(msg string) {
	cast.LogErr(fmt.Sprintf("DBG: %s", msg), d.runtime.irb.ExecutionUUID)
}

func (d *debugger) lw(msg string) {
	cast.LogWarn(fmt.Sprintf("DBG: %s", msg), d.runtime.irb.ExecutionUUID)
}

func (d *debugger) Detach(actionMFD io.ReadWriteCloser) {
	vpty := nsterm.NewVirtPTY()
	ldisc := nsterm.NewMultiUserLdisc()
	// ldisc := nsterm.NewRawLdisc()
	vpty.SetLDisc(ldisc)
	d.detach = append(d.detach, vpty)
	mfd := vpty.MustarFD()
	sfd := vpty.SluvaFD()

	go func() {
		_, _ = io.Copy(sfd, actionMFD)
		// fmt.Println("out of io.Copy(sfd, actionMFD)")
		mfd.Close()
		vpty.Close()
	}()
	go func() {
		_, _ = io.Copy(actionMFD, sfd)
		// fmt.Println("out of io.Copy(actionMFD, sfd)")
		sfd.Close()
		vpty.Close()
	}()
}

func (d *debugger) Attach(clientFD io.ReadWriteCloser, num int) error {
	// cc.Raw(true)
	// defer cc.Raw(false)
	// clientFD := cc.GetFD() // sluva
	if num < 0 || len(d.detach) == num {
		return fmt.Errorf("attaching to unknown term num")
	}
	vpty := d.detach[num]

	nmp := vpty.NewMustarPort()
	// this is a multiuser ldisc, doing a cursor
	// forces ldisc to know this port
	vpty.CursorMustar(nmp)
	mOutFD := nmp.OutFD()

	// mfd := vpty.MustarFD()
	// mfd.Write([]byte(nebulant_term.IdentifyDevice))

	go func() {
		_, _ = io.Copy(clientFD, mOutFD)
		// fmt.Println("out of atachment io.Copy(clientFD, mOutFD)")
		mOutFD.Close()
	}()
	_, _ = io.Copy(mOutFD, clientFD)
	// fmt.Println("out of atachment io.Copy(mfd, clientFD)")

	// go func() {
	// 	_, _ = io.Copy(actionMustarFD, clientFD)
	// 	fmt.Println("\n\n\nout of attach clientfd -> f\n\n\n")

	// }()
	// // copy os pty to sluva
	// _, err := io.Copy(clientFD, actionMustarFD)
	// fmt.Println("\n\n\nout of attach f -> clientfd\n\n\n")
	// clientFD.Write([]byte("attach finished\n"))
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (d *debugger) printDebugBannerStart() {
	tit := "debug session START event"
	fill := strings.Repeat("/", (80-len(tit))/2)
	d.l(fmt.Sprintf("%s %s %s", fill, tit, fill))
}
func (d *debugger) printDebugBannerEnd() {
	d.l(strings.Repeat("\\", 80))
	d.l("debug session END event")
	d.l(strings.Repeat("\\", 80))
}

func (d *debugger) Start() {
	d.printDebugBannerStart()
	if d.runtime.IsServerMode() {
		d.l("Starting debugger in server mode...")
		d.startServerMode(make(chan struct{}))
	} else {
		d.l("Starting debugger in local shell...")
		// this could still start server mode
		// if term.IsTerminal returns false
		d.startLocalTermMode()
	}
	d.printDebugBannerEnd()
}

// no server mode
func (d *debugger) startLocalTermMode() {
	d.running = true
	oSstdin := nebulant_term.GenuineOsStdin

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		d.lw("No terminal detected. Remote debugger system will start. You can still press any key to start local shell")
		started := make(chan struct{})
		go d.startServerMode(started)
		<-started
		go d.ReverseServer()
		b := make([]byte, 1)
		oSstdin.Read(b)
	}

	lfd := nsterm.NewFD("local term", nebulant_term.GenuineOsStdin, nebulant_term.GenuineOsStdout)
	clnt := &client{
		conn: nil,
		dbg:  d,
		vpty: nsterm.NewVirtPTY(),
		wsrw: lfd,
		space: &clientSpace{
			store: storage.NewStore(),
		},
	}
	d.lclient = clnt
	d.lclient.MakeRealOSTermRaw()
	defer d.lclient.RestorerealOSTerm()
	clnt.start(len(d.detach) > 0)
}

func (d *debugger) startServerMode(started chan struct{}) error {
	d.running = true
	d.remote = true
	// TODO: lookup for available port?
	id := d.runtime.irb.ExecutionUUID
	srv := nhttpd.GetServer()
	srv.AddView(`/debugger/`+*id, d.debuggerView)
	d.close = srv.ServeIfNot()
	scheme := "ws"
	if srv.GetScheme() == "https" {
		scheme = "wss"
	}
	d.l("New local debugger started")
	d.l(fmt.Sprintf("Use `nebulant debugger %s://%s/debugger/%s` to get attached to local debug session", scheme, srv.GetAddr(), *id))
	started <- struct{}{}
	err := <-d.close
	d.remote = false
	// WIP: una forma de apagar el server aquí, del mismo
	// modo que se inicializa: StopAsNeeded(), quizás si
	// ServeIfNot devuelve un "subscription" que se pasa
	// al cerrar, o quzás si ServeIfNot cuenta los arranques
	// y los a pagados, cuando sea 0, se apaga
	d.l("Remote debugger out")
	return err
}

func (d *debugger) _bridgeConnect(newp *newBridgePoolSerializer) error {
	u := newp.cliURL
	token := newp.Token
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else if u.Scheme == "http" {
		u.Scheme = "ws"
	} // else: malformed url on Dial
	// u.Path = u.Path + "/cli"

	d.l(fmt.Sprintf("Remote debug session token: %s", token))
	d.l(fmt.Sprintf("Dialing the bridge... %s", u.String()))

	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Basic %s", token))
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}
	c, _, err := dialer.Dial(u.String(), headers)
	if err != nil {
		return errors.Join(fmt.Errorf("dialing bridge err"), err)
	}
	defer c.Close()
	wsrw := ws.NewWebSocketReadWriteCloser(c)

	clnt := &client{
		conn: nil,
		dbg:  d,
		vpty: nsterm.NewVirtPTY(),
		wsrw: wsrw,
		space: &clientSpace{
			store: storage.NewStore(),
		},
	}

	d.l("Connection to the bridge stablished")
	d.l("Now you can connect to the debug session using the cli or a web browser:")
	u.Path = newp.ConsumerPath
	d.l(fmt.Sprintf("With cli: ./nebulant debugger %s", u.String()))

	u.Path = newp.XtermPath
	if u.Scheme == "wss" {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}
	d.l(fmt.Sprintf("With web browser: %s", u.String()))

	clnt.start(len(d.detach) > 0)
	return nil
}

func (d *debugger) reverseCloudServer() error {
	url := &url.URL{
		Scheme: config.BASE_SCHEME,
		Host:   config.BACKEND_API_HOST,
		Path:   "v1/bridge/new/",
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}
	jar, err := config.Login(nil)
	if err != nil {
		return errors.Join(fmt.Errorf("backend login problem"), err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Jar: jar}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rawbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 399 {
		return fmt.Errorf(strconv.Itoa(resp.StatusCode) + " server error: " + string(rawbody))
	}

	newp := &newBridgePoolSerializer{}
	if err := json.Unmarshal(rawbody, newp); err != nil {
		return err
	}

	url.Host = net.JoinHostPort(newp.Host, strconv.Itoa(newp.Port))
	url.Path = newp.CliPath
	newp.cliURL = url
	return d._bridgeConnect(newp)
}

func (d *debugger) reverseSelfServer() error {
	u, err := url.Parse(*config.BridgeAddrFlag)
	if err != nil {
		return err
	}
	if u.Host == "" {
		h, p, err := net.SplitHostPort(*config.BridgeAddrFlag)
		if err != nil {
			return err
		}
		u.Scheme = "https"
		u.Host = net.JoinHostPort(h, p)
	}
	h, sp, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}
	p, err := strconv.Atoi(sp)
	if err != nil {
		return err
	}
	if u.Path == "" {
		u.Path = "/new"
	}

	rawBody, err := json.Marshal(map[string]string{
		"auth": *config.BridgeSecretFlag,
	})
	reqBody := bytes.NewBuffer(rawBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u.String(), reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	httpcl := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}}
	resp, err := httpcl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return fmt.Errorf("HTTP ERR: status code %v", resp.StatusCode)
	}
	rawbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	newPool := make(map[string]string)
	if err := json.Unmarshal(rawbody, &newPool); err != nil {
		return err
	}
	if _, exists := newPool["token"]; !exists {
		return fmt.Errorf("conn err: no token provided")
	}

	u.Path = "/cli"

	newp := &newBridgePoolSerializer{
		Token:        newPool["token"],
		Scheme:       u.Scheme,
		Host:         h,
		Port:         p,
		ConsumerPath: fmt.Sprintf("/consumer/%s", newPool["token"]),
		XtermPath:    fmt.Sprintf("/xterm/%s", newPool["token"]),
		CliPath:      "/cli",
		cliURL:       u,
	}
	return d._bridgeConnect(newp)
}

// use bridge
func (d *debugger) ReverseServer() {
	d.l("A new remote debug session across bridge will be started")
	if *config.BridgeAddrFlag != "" {
		// self-hosted bridge
		d.l("Dialing self-hosted bridge...")
		err := d.reverseSelfServer()
		if err != nil {
			cast.LogErr(err.Error(), nil)
		}
		return
	}
	d.l("Dialing nebulant cloud bridge...")
	err := d.reverseCloudServer()
	if err != nil {
		cast.LogErr(err.Error(), nil)
	}
}

func (d *debugger) closeClient(cc *client) {
	if d.lclient == cc {
		// local client
		d.lclient = nil
	} else {
		delete(d.rclients, cc)
	}

	if d.lclient == nil && len(d.rclients) <= 0 {
		if d.remote {
			d.close <- nil
		}
		d.l("no more clients in debugger, exiting")
		d.running = false
		d.runtime.Play() // resume run on last debugger out
		delete(debuggers, d.runtime)
	}
}

func (d *debugger) debuggerView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	var upgrader = websocket.Upgrader{
		// ReadBufferSize:  MAXREADSIZE,
		// WriteBufferSize: MAXWRITESIZE,
		CheckOrigin: func(r *http.Request) bool { return true },
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
		space: &clientSpace{
			store: storage.NewStore(),
		},
	}

	d.rclients[clnt] = true
	go clnt.start(len(d.detach) > 0)
}

///// WIP: mover estas funcs a algún otro sitio

type printBoxConf struct {
	prefix   string
	sufix    string
	linkline bool
}

var regularBox *printBoxConf = &printBoxConf{
	prefix:   "└─",
	sufix:    "─┐",
	linkline: true,
}

var forkItemBox *printBoxConf = &printBoxConf{
	prefix:   "├─",
	sufix:    "─*",
	linkline: false,
}

var forkLastItemBox *printBoxConf = &printBoxConf{
	prefix:   "└─",
	sufix:    "─*",
	linkline: false,
}

// _printBox
// fd -> file descriptor to write
// actx -> the actx of the action to write
// ponter -> if should draw ponter to run state
// end -> if this box has no childrens: prevent draw output arrow
// brother -> if this child is part of others: draw left-lines
// WIP: mejorar esto
func _printBox(fd io.Writer, actx base.IActionContext, pointer bool, cfg *printBoxConf) {
	aname := "null"
	action := actx.GetAction()
	aname = action.ActionName
	runstatus := actx.GetRunStatus()

	statustxt := ""
	switch runstatus {
	case base.RunStatusArranging:
		statustxt = "arranging (one step more to run)"
	case base.RunStatusRunning:
		statustxt = "running"
	case base.RunStatusDone:
		statustxt = "done"
	}

	// print actx
	if pointer && (runstatus > base.RunStatusReady && runstatus <= base.RunStatusRunning) {
		fmt.Fprintf(fd, "═>%s%s%s\t%p %s\n", cfg.prefix, aname, cfg.sufix, actx, statustxt)
	} else {
		fmt.Fprintf(fd, "  %s%s%s\t%p %s\n", cfg.prefix, aname, cfg.sufix, actx, statustxt)
	}

	if !cfg.linkline {
		return
	}
	// print output lines
	if pointer && runstatus > base.RunStatusRunning {
		fmt.Fprintf(fd, "═>┌─%s─┘\n", strings.Repeat("─", len(aname)))
	} else {
		fmt.Fprintf(fd, "  ┌─%s─┘\n", strings.Repeat("─", len(aname)))
	}

}

func printActxTrace(fd io.Writer, actx base.IActionContext, pointer bool) {
	fmt.Fprintln(fd, "  │")
	parents := actx.Parents()
	for _, parent := range parents {
		parents2 := parent.Parents()
		for _, parent2 := range parents2 {
			_printBox(fd, parent2, false, regularBox)
		}
		_printBox(fd, parent, false, regularBox)
	}
	// action := actx.GetAction()

	_printBox(fd, actx, pointer, regularBox)

	// if actx.GetRunStatus() < base.RunStatusRunning {
	// 	fmt.Fprintln(fd, "->■")
	// }
	// if actx.GetRunStatus() == base.RunStatusRunning {
	// 	fmt.Fprintf(fd, "->└%s\t%p\n", action.ActionName, actx)
	// 	fmt.Fprint(fd, "  ┌────────┘\n")
	// } else {
	// 	fmt.Fprintf(fd, "  └%s\t%p\n", action.ActionName, actx)
	// 	fmt.Fprint(fd, "  ┌────────┘\n")
	// }
	// if actx.GetRunStatus() > base.RunStatusRunning {
	// 	fmt.Fprintln(fd, "->■")
	// }

	children := actx.Children()
	size := len(children)
	count := 0
	if size <= 0 {
		return
	}
	if size > 1 {
		for _, child := range children {
			count++
			if size == count {
				_printBox(fd, child, false, forkLastItemBox)
				return
			}
			_printBox(fd, child, false, forkItemBox)
		}
		return
	}
	_printBox(fd, children[0], false, regularBox)
}

////

func (d *debugger) hightLightCursor() {
	if d.cursor == nil {
		return
	}
	action := d.cursor.GetAction()
	active_ids := d.runtime.activeActionsID.Slice()
	active_ids = append(active_ids, action.ActionID)
	cast.PushState(active_ids, cast.EventRuntimeStarted, d.runtime.irb.ExecutionUUID)
}

func (d *debugger) WriteToAllClients(txt string) {
	for cc := range d.rclients {
		fmt.Fprint(cc.GetFD(), txt)
	}
	if d.lclient != nil {
		fmt.Fprint(d.lclient.GetFD(), txt)
	}
}

func (d *debugger) Wait() {
	d.waiting = true
	defer func() { d.waiting = false }()
	d.WriteToAllClients("continue normal run...\n")
	event := d.elistener.WaitUntil([]base.EventCode{
		base.RuntimeEndEvent,
		base.DebugOnEvent,
	})
	if event == base.RuntimeEndEvent {
		d.WriteToAllClients("bp run finished, bye.\n")
		// close all clients
		for cc := range d.rclients {
			go cc.close()
		}
		if d.lclient != nil {
			d.lclient.close()
		}
	}
	d.WriteToAllClients("New breakpoint\n")
}

func (d *debugger) ExecCmd(cc *client, cmd string) {
	cast.LogDebug(fmt.Sprintf("processing command %s", cmd), nil)
	clientFD := cc.GetFD()
	if d.waiting {
		fmt.Fprint(clientFD, "waiting next breakpoint...\n")
		return
	}

	argv, err := shlex.Split(cmd)
	if err != nil {
		fmt.Fprintf(clientFD, "cmd err: %s\n", err.Error())
	}
	fs := flag.NewFlagSet("dbg", flag.ContinueOnError)
	fs.Parse(argv)

	switch string(fs.Arg(0)) {
	case "nebu!":
		fmt.Fprintf(clientFD, "lant!\n")
	case "h", "help", "hh", "ayuda", "ajuda", "jelp", "jalp", "aiuda", "aiudapremo", "?":
		fmt.Fprintf(clientFD, `
	h - This help txt ;)
	j - Jump to another thread
	n - Runs the next step
	ll - Prints current cursor position
	set - Sets new vars values
	p var - Prints the value of the variable "var" 
	u - Goes up to previous step 
	desc - Prints thread/action properties
	th - Prints current thread
	shell - Open new local shell
	c - Continues running the blueprint
	q - Quits debugger. Have a nice day.
	exit - Like q. You are a nice person.

`)
	case "c":
		d.runtime.Play()
		d.cursor = nil
		d.Wait()
	case "ll":
		printActxTrace(clientFD, d.cursor, true)
	// case "hl":
	// 	if fs.Arg(1) == "" {

	// 	}
	case "s":
		fmt.Fprint(clientFD, "Step not implemented yet\n")
	case "j":
		if fs.Arg(1) == "" {
			if d.cursor == nil || !d.cursor.IsThreadPoint() {
				threads := d.runtime.GetThreads()
				for th := range threads {
					curr := th.GetCurrent()
					if curr != d.cursor {
						d.SetCursor(curr)
						fmt.Fprintf(clientFD, "joining into the thread %p with action %p\n", th, curr)
						return
					}
				}
				fmt.Fprint(clientFD, "err: cannot join into thread. Sugestion: j 0xThreadID\n")
				return
			}
			for _, chld := range d.cursor.Children() {
				threads := d.runtime.GetThreads()
				for th := range threads {
					if th.hasActionContext(chld) {
						curr := th.GetCurrent()
						if curr != nil {
							d.SetCursor(curr)
							fmt.Fprintf(clientFD, "joining into the thread %p with action %p\n", th, curr)
							return
						}
					}
				}
			}
			return
		}
		threads := d.runtime.GetThreads()
		arg1 := fs.Arg(1)
		for th := range threads {
			if fmt.Sprintf("%p", th) == arg1 {
				curr := th.GetCurrent()
				if curr != nil {
					d.SetCursor(curr)
					fmt.Fprintf(clientFD, "joining into the thread %p with action %p\n", th, curr)
					return
				}
				fmt.Fprint(clientFD, "err: this thread has no actions\n")
				return
			}
			fmt.Fprint(clientFD, "err: thread not found\n")
		}
	case "n":
		threads := d.runtime.GetThreads()
		for th := range threads {
			if th.hasActionContext(d.cursor) {
				fmt.Fprintf(clientFD, "thread %p has %p\n", th, d.cursor)
				prevcurrent := th.GetCurrent()
				// if prevcurrent.IsThreadPoint() {
				// 	// check if there is already running thread that has children
				// 	chldfound := false
				// 	for _, chld := range prevcurrent.Children() {
				// 		if th.hasActionContext(chld) {
				// 			chldfound = true
				// 			break
				// 		}
				// 	}
				// 	if chldfound {
				// 		fmt.Fprint(clientFD, "Be carefully with forky!!!!\n")
				// 		fmt.Fprintf(clientFD, "There is already a thread(s) running children.")
				// 		fmt.Fprintf(clientFD, "Use j to jump to already running action.")
				// 		fmt.Fprintf(clientFD, "If you continue with n, new threads")
				// 		fmt.Fprintf(clientFD, "will be created.")
				// 	}
				// }
				sconfirm, ok := th.Step()
				if !ok {
					fmt.Fprint(clientFD, "cannot go next\n")
					return
				}
				<-sconfirm
				nextcurrent := th.GetCurrent()
				if nextcurrent == nil {
					// no more actions in this thread
					fmt.Fprint(clientFD, "No more actions in this thread\n")
					if prevcurrent == nil {
						fmt.Fprint(clientFD, "prevcurrent\n")
						return
					}

					actxs := prevcurrent.Children()
					count := len(actxs)
					if count > 0 {
						d.SetCursor(actxs[0])
						if count > 1 {
							fmt.Fprintf(clientFD, "New threads (%v) created. Joining into the first of them\n", count)
						}
					}
					return
				}
				d.SetCursor(nextcurrent)
				return
			}
		}
		fmt.Fprintf(clientFD, "cannot found thread of %p\n", d.cursor)
	case "hl":
		d.hightLightCursor()
	case "set":
		fmt.Fprint(clientFD, fs.Args())
		args := fs.Args()[1:]
		if len(args) < 3 || args[1] != "=" {
			fmt.Fprintf(clientFD, "invalid syntax\n")
			return
		}
		cc.space.store.SetPrivateVar(args[0], args[2])
	case "p":
		if fs.Arg(1) == "" {
			// print all contexts
			for _, actx := range d.runtime.GetStack() {
				action := actx.GetAction()
				if action.Output != nil {
					fmt.Fprintf(clientFD, "- name:%s \t : actx:%p (outputname:%s) (store:%p)\n", action.ActionName, actx, *action.Output, actx.GetStore())
					continue
				}
				fmt.Fprintf(clientFD, "- name:%s \t : actx:%p (outputname:null) (store:%p)\n", action.ActionName, actx, actx.GetStore())
			}
			return
		}

		path := fs.Arg(1)

		pv := cc.space.store.GetPrivateVar(path)
		if pv != nil {
			fmt.Fprintf(clientFD, "%s", pv)
			fmt.Fprintf(clientFD, "\n")
			return
		}

		aa := strings.Split(path, ".")
		pv = cc.space.store.GetPrivateVar(aa[0])
		if pv != nil {
			aa[0] = fmt.Sprintf("%s", pv)
		}

		for _, actx := range d.runtime.GetStack() {
			if aa[0] == fmt.Sprintf("%p", actx) {
				actxstore := actx.GetStore()
				action := actx.GetAction()
				refname := action.Output
				if refname == nil {
					fmt.Fprintf(clientFD, "action without data\n")
					return
				}
				aa[0] = *refname
				ppp := fmt.Sprintf("{{%s}}", strings.Join(aa, "."))
				err := actxstore.Interpolate(&ppp)
				if err != nil {
					fmt.Fprint(clientFD, err.Error())
					fmt.Fprintf(clientFD, "\n")
					return
				}
				fmt.Fprint(clientFD, ppp)
				fmt.Fprintf(clientFD, "\n")
				return
			}
			store := actx.GetStore()
			if aa[0] == fmt.Sprintf("%p", store) {
				v, err := store.GetRawJSONValues()
				if err != nil {
					fmt.Fprint(clientFD, err.Error())
					fmt.Fprintf(clientFD, "\n")
					return
				}
				enc, err := json.MarshalIndent(v, "", "    ")
				if err != nil {
					fmt.Fprint(clientFD, err.Error())
					fmt.Fprintf(clientFD, "\n")
					return
				}
				fmt.Fprint(clientFD, "storage dump:")
				fmt.Fprintf(clientFD, "%s\n", enc)
				return
			}
		}

		actxstore := d.cursor.GetStore()
		ppp := fmt.Sprintf("{{%s}}", strings.Join(aa, "."))
		actxstore.Interpolate(&ppp)
		fmt.Fprint(clientFD, ppp)
		fmt.Fprintf(clientFD, "\n")
	case "u":
		parents := d.cursor.Parents()
		if len(parents) == 1 && parents[0].IsThreadPoint() {
			d.runtime.NewThread(parents[0])
			fmt.Fprintf(clientFD, "Fork point. New thread created.\n")
			d.SetCursor(parents[0])
			return
		}

		threads := d.runtime.GetThreads()
		for th := range threads {
			if th.hasActionContext(d.cursor) {
				fmt.Fprintf(clientFD, "thread %p has %p\n", th, d.cursor)
				bconfirm, ok := th.StackUp()
				if !ok {
					fmt.Fprint(clientFD, "cannot go back\n")
					return
				}
				<-bconfirm
				prevcurrent := th.GetCurrent()
				if prevcurrent == nil {
					// no more actions in this thread
					fmt.Fprint(clientFD, "No more actions back\n")
					return
				}
				d.SetCursor(prevcurrent)
				return
			}
		}
		fmt.Fprintf(clientFD, "cannot found thread of %p\n", d.cursor)
	case "desc":
		if fs.Arg(1) == "" {
			if d.cursor == nil {
				return
			}
			threads := d.runtime.GetThreads()
			for th := range threads {
				if th.hasActionContext(d.cursor) {
					fmt.Fprintf(clientFD, "Thread %p:\n", th)
					state := th.GetState()
					statetxt := ""
					switch state {
					case base.RuntimeStatePlay:
						statetxt = "running"
					case base.RuntimeStateStill:
						statetxt = "paused"
					case base.RuntimeStateEnd:
						statetxt = "done"
					}
					fmt.Fprintf(clientFD, "\tState: %s\n", statetxt)
					fmt.Fprintf(clientFD, "\tCurrent: %p\n", th.GetCurrent())
					fmt.Fprintf(clientFD, "\tQueue len: %v\n", len(th.GetQueue()))
					fmt.Fprintf(clientFD, "\tDone len: %v\n", len(th.done))
					fmt.Fprintf(clientFD, "\tExitCode: %v\n", th.ExitCode)
					fmt.Fprintf(clientFD, "\tExitErr: %v\n", th.ExitErr)
					fmt.Fprintf(clientFD, "\tEvent chan len: %v\n", len(th.EventListener().EventChan()))
					//
					fmt.Fprintf(clientFD, "Action: %p:\n", d.cursor)
					statetxt = ""
					switch d.cursor.GetRunStatus() {
					case base.RunStatusReady:
						statetxt = "ready"
					case base.RunStatusArranging:
						statetxt = "arranging"
					case base.RunStatusRunning:
						statetxt = "running"
					case base.RunStatusDone:
						statetxt = "done"
					}
					fmt.Fprintf(clientFD, "\tState: %s\n", statetxt)
					fmt.Fprintf(clientFD, "\tType: %v\n", d.cursor.Type())
					fmt.Fprintf(clientFD, "\tName: %v\n", d.cursor.GetAction().ActionName)
					fmt.Fprintf(clientFD, "\tParents: %v\n", d.cursor.Parents())
					fmt.Fprintf(clientFD, "\tChildren: %v\n", d.cursor.Children())
					fmt.Fprintf(clientFD, "\tStore: %p\n", d.cursor.GetStore())
					action := d.cursor.GetAction()
					if action != nil {
						var prettyJSON bytes.Buffer
						err = json.Indent(&prettyJSON, action.Parameters, "", "    ")
						if err != nil {
							prettyJSON.Write([]byte(err.Error()))

						}
						fmt.Fprintf(clientFD, "\tParameters: %s\n", prettyJSON.String())
					}
				}
			}
			return
		}
		fs.Arg(1)
		threads := d.runtime.GetThreads()
		for th := range threads {
			fmt.Fprintf(clientFD, "comparing %p against %s\n", th, fs.Arg(1))
			if fmt.Sprintf("%p", th) == fs.Arg(1) {
				state := th.GetState()
				statetxt := ""
				switch state {
				case base.RuntimeStatePlay:
					statetxt = "running"
				case base.RuntimeStateStill:
					statetxt = "paused"
				case base.RuntimeStateEnd:
					statetxt = "done"
				}

				fmt.Fprint(clientFD, "thread found\n")
				fmt.Fprintf(clientFD, "State: %s\n", statetxt)
				fmt.Fprintf(clientFD, "current: %p\n", th.GetCurrent())
				fmt.Fprintf(clientFD, "queue len: %v\n", len(th.GetQueue()))
				fmt.Fprintf(clientFD, "done len: %v\n", len(th.done))
				fmt.Fprintf(clientFD, "ExitCode: %v\n", th.ExitCode)
				fmt.Fprintf(clientFD, "ExitErr: %v\n", th.ExitErr)
				fmt.Fprintf(clientFD, "event chan len: %v\n", len(th.EventListener().EventChan()))
			}
		}
	case "th":
		threads := d.runtime.GetThreads()
		for th := range threads {
			fmt.Fprintf(clientFD, " thread %p\n", th)
			curr := th.GetCurrent()
			if curr == nil {
				continue
			}
			printActxTrace(clientFD, curr, d.cursor == curr)
			fmt.Println("")
		}
	case "shell":
		// cc.stoppipe = make(chan struct{})
		// defer close(cc.stoppipe)

		var shell string
		switch fs.Arg(1) {
		case "wsl":
			if runtime.GOOS != "windows" {
				fmt.Fprintln(clientFD, fmt.Errorf("wsl support just in windows"))
				return
			}
			shell = "cmd.exe /c wsl.exe"
		case "powershell", "ps":
			if runtime.GOOS != "windows" {
				fmt.Fprintln(clientFD, fmt.Errorf("powershell support just in windows"))
				return
			}
			shell = "powershell.exe"
		case "":
			shell, err = nebulant_term.DetermineOsShell()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				fmt.Fprintln(clientFD, err.Error())
				return
			}
			if strings.HasSuffix(shell, "cmd.exe") {
				shell = shell + " /T:0D /U /E:ON"
			}
		default:
			shell = fs.Arg(1)
		}

		cc.RestorerealOSTerm()

		f, err := nebulant_term.GetOSPTY(&nebulant_term.OSPTYConf{Shell: shell})
		if err != nil {
			cast.LogErr(err.Error(), d.runtime.irb.ExecutionUUID)
			fmt.Fprintln(clientFD, err.Error())
			return
		}
		defer f.Close()

		cc.Raw(true)
		defer cc.Raw(false)
		cc.MakeRealOSTermRaw()

		fmt.Fprintf(clientFD, "Running shell %v...", shell)

		// copy sluva to os pty
		go func() {
			_, _ = io.Copy(f, clientFD)
		}()
		// copy os pty to sluva
		go func() {
			fmt.Fprintf(clientFD, "%s\n", nebulant_term.Reset)
			fmt.Fprintf(clientFD, "%s\n", nebulant_term.EraseDisplay)
			fmt.Fprintf(clientFD, "%s\n", nebulant_term.RestoreCursor)
			_, _ = io.Copy(clientFD, f)
		}()

		exitCode, err := f.Wait(context.Background())
		fmt.Fprintf(clientFD, "%s\n", nebulant_term.Reset)
		fmt.Fprintf(clientFD, "%s\n", nebulant_term.EraseDisplay)
		fmt.Fprintf(clientFD, "%s\n", nebulant_term.RestoreCursor)
		if err != nil {
			fmt.Fprintln(clientFD, err.Error())
			cast.LogErr(err.Error(), d.runtime.irb.ExecutionUUID)
		}

		fmt.Fprintf(clientFD, "local shell stopped with exit code %v\n", exitCode)

	default:
		cast.LogDebug("Unknown command "+string(cmd), nil)
		fmt.Fprintln(clientFD, "Unknown command "+string(cmd))
	}
}

type client struct {
	dbg  *debugger
	conn *websocket.Conn // in case of remote clients
	wsrw io.ReadWriteCloser
	vpty *nsterm.VPTY2
	// default ldisc, used in Raw(false)
	ldisc   nsterm.Ldisc
	sluvaFD *nsterm.PortFD
	space   *clientSpace
	// state of term on init. Used to
	// restore to original state
	termInitState *term.State
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

func (c *client) IsLocalTerm() bool {
	r := c.vpty.MustarFD().GetRawR()
	_, ok := r.(*os.File)
	fmt.Printf("is local term? %v\n", ok)
	return ok
}

func (c *client) MakeRealOSTermRaw() {
	fmt.Println("setting os real RAW state")
	stdin := nebulant_term.GenuineOsStdin
	// in localterm the local term should set as raw
	// to leave VPTY handle raw data
	oldState, err := term.MakeRaw(int(stdin.Fd()))
	if err != nil {
		panic(err)
	}
	if c.termInitState == nil {
		c.termInitState = oldState
	}
}

func (c *client) RestorerealOSTerm() {
	fmt.Println("restoring os term state")
	if c.termInitState == nil {
		return
	}
	stdin := nebulant_term.GenuineOsStdin
	term.Restore(int(stdin.Fd()), c.termInitState)
}

func (c *client) close() {
	c.conn.Close()
}

func (c *client) start(attach bool) {
	ldisc := nsterm.NewDefaultLdisc()
	c.vpty.SetLDisc(ldisc)
	c.ldisc = ldisc

	// for wsock client
	// stdin and stdout for xterm.js
	mfd := c.vpty.MustarFD()
	go func() {
		// copy app rcv to mustar file descriptor
		_, err := io.Copy(mfd, c.wsrw)
		if err != nil {
			fmt.Println(err.Error())
		}
		// fmt.Println("out of wsrw to mfd")
	}()
	go func() {
		// copy mustar file descriptor input to app in interface
		io.Copy(c.wsrw, mfd)
		// fmt.Println("out of mfd wsrw")
	}()

	// file descriptor for app
	sfd := c.vpty.SluvaFD()
	c.sluvaFD = sfd

	// welcome msg to debug session
	sfd.Write([]byte(nebulant_term.Magenta + "Nebulant debug. Hello :)\r\nhow are you?\r\n" + nebulant_term.Reset))

	// if there is a detached vpty iface, attach to it
	if attach {
		nsluva := c.vpty.NewSluvaPort()
		c.vpty.CursorSluva(nsluva)
		c.Raw(true)
		sfd.Write([]byte("Attaching to action..."))
		c.dbg.Attach(nsluva.OutFD(), 0)
		c.vpty.DestroyPort(nsluva)
		c.Raw(false)
		// no restoration needed, there is
		// only one sluva in vpty after
		// attach: the original sluva
		// c.vpty.CursorSluva(...)
		// fmt.Println("out of attach")
	}

	prmpt := nsterm.NewPrompt(c.vpty, sfd, sfd)
	prmpt.SetPS1("Nebulant dbg> ")
	for {
		cursor := c.dbg.GetCursor()
		if cursor != nil {
			action := cursor.GetAction()
			prmpt.SetPS1(fmt.Sprintf("Nebulant dbg [%s  %p]> ", action.ActionName, cursor))
		}

		s, err := prmpt.ReadLine()
		if err != nil {
			// TODO: disconnect with err
		}
		if s == nil {
			// no command
			continue
		}

		if *s == "term restore" {
			c.RestorerealOSTerm()
			continue
		}

		if *s == "term raw" {
			c.MakeRealOSTermRaw()
			continue
		}

		// built in :)
		if *s == "exit" || *s == "q" {
			c.dbg.closeClient(c)
			// err := mfd.Close()
			// if err != nil {
			// 	cast.LogErr(err.Error(), c.dbg.runtime.irb.ExecutionUUID)
			// }
			sfd.Write([]byte("Closing debugger client..."))
			if c.conn != nil {
				sfd.Write([]byte("Closing remote connection..."))
				err := c.conn.Close()
				if err != nil {
					c.dbg.le(err.Error())
				}
			}
			c.dbg.l("dbg client out")
			break
		}

		// built in ;)
		if *s == "help" {
			// show debug help
		}

		c.dbg.ExecCmd(c, *s)
		c.dbg.hightLightCursor()
	}
}
