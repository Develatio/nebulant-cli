// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

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

package subcom

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	ws "github.com/develatio/nebulant-cli/netproto/websocket"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

var MAXWRITESIZE = 1024
var MAXREADSIZE = 1024

// type ioWSrw struct {
// 	conn *websocket.Conn
// }

// func (i *ioWSrw) Write(p []byte) (int, error) {
// 	if len(p) > MAXREADSIZE {
// 		err := i.conn.WriteMessage(websocket.BinaryMessage, p[:MAXREADSIZE-1])
// 		if err != nil {
// 			return MAXREADSIZE, err
// 		}
// 		n, err := i.Write(p[MAXREADSIZE:])
// 		if err != nil {
// 			return MAXREADSIZE + n, err
// 		}
// 		return MAXREADSIZE + n, nil
// 	}
// 	err := i.conn.WriteMessage(websocket.BinaryMessage, p)
// 	if err != nil {
// 		return len(p), err
// 	}
// 	return len(p), nil

// }

// func (i *ioWSrw) Read(p []byte) (n int, err error) {
// 	_, m, err := i.conn.ReadMessage()
// 	if err != nil {
// 		return 0, err
// 	}

// 	ml := len(m)
// 	pl := len(p)
// 	fl := ml
// 	if ml > pl {
// 		fl = pl
// 	}
// 	for i := 0; i < fl; i++ {
// 		p[i] = m[i]
// 	}

// 	return len(p), nil
// }

func parseDebuggFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("debugger", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant debugger <addr>\n")
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func DebuggerCmd(nblc *subsystem.NBLcommand) (int, error) {
	cmdline := nblc.CommandLine()
	fs, err := parseDebuggFs(cmdline)
	if err != nil {
		return 1, err
	}
	// ver si podemos exponer esta info en el sistema de envs
	// para extraerlo automáticamente sin necesidad
	// de indicar el host
	addr := cmdline.Arg(1)
	if len(addr) <= 0 {
		fs.Usage()
		return 1, fmt.Errorf("please, provide addr")
	}
	addr = strings.ToLower(addr)
	addr, _ = strings.CutPrefix(addr, "ws://")
	addr, _ = strings.CutPrefix(addr, "http://")
	addr, _ = strings.CutPrefix(addr, "https://")
	addr, _ = strings.CutPrefix(addr, "wss://")

	u, err := url.Parse(fmt.Sprintf("ws://%s", addr))
	if err != nil {
		fs.Usage()
		return 1, err
	}
	u.Scheme = "ws"
	// u := url.URL{Scheme: "ws", Host: uu.Host, Path: ""}
	log.Printf("connecting to %s", u.String())

	// WIP: ver cómo determinar esto
	headers := make(http.Header)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return 1, err
	}
	c.SetReadLimit(int64(MAXREADSIZE))
	defer c.Close()

	// 	done := make(chan struct{})

	// 	go func() {
	// 		defer close(done)
	// 		for {
	// 			_, message, err := c.ReadMessage()
	// 			if err != nil {
	// 				log.Println("read:", err)
	// 				return
	// 			}
	// 			// term.Print(message)
	// 			// log uses term.Output (and can handle multiline)
	// 			fmt.Printf("%s", message)
	// 		}
	// 	}()

	// 	counter2 := 0
	// 	// lin := term.AppendLine()
	// 	// defer lin.Close()

	//L:
	// fmt.Println("looping")
	// // raw term

	if _, isFD := nblc.Stdin.(*os.File); isFD {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(0, oldState)
	}

	// os.Setenv("TERM", "vt100")

	wsrw := ws.NewWebSocketReadWriteCloser(c)

	go io.Copy(nblc.Stdout, wsrw)
	io.Copy(wsrw, nblc.Stdin)
	return 0, nil

	// go func() {
	// 	// var bff []byte
	// 	// var buff = bytes.NewBuffer(bff)
	// 	reader := bufio.NewReader(os.Stdin)
	// 	for {
	// 		char, _, err := reader.ReadRune()
	// 		if err != nil {
	// 			fmt.Println("ERRRRRR 1")
	// 			return
	// 		}
	// 		// buff.WriteRune(char)
	// 		// fmt.Println("writing to iows", char)
	// 		// fmt.Printf("Rune read: %d", char)
	// 		// fmt.Println("----")
	// 		iows.Write([]byte(string(char)))
	// 	}

	// }()

	// for {
	// 	// bff := make([]byte, MAXREADSIZE)
	// 	// _, err := iows.Read(bff)
	// 	// if err != nil {
	// 	// 	return 1, err
	// 	// }
	// 	// fmt.Println(bff)
	// 	// fmt.Printf("v:%v d:%d s:%s", bff, bff, bff)
	// 	// fmt.Printf("%s", bff)
	// 	io.Copy(os.Stdout, iows)
	// 	// switch buff.String() {
	// 	// case "exit":
	// 	// 	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 	// 	if err != nil {
	// 	// 		log.Println("write close:", err)
	// 	// 		return 1, err
	// 	// 	}
	// 	// 	select {
	// 	// 	case <-done:
	// 	// 	case <-time.After(time.Second):
	// 	// 	}
	// 	// 	break L
	// 	// default:
	// 	// 	err := c.WriteMessage(websocket.BinaryMessage, buff.Bytes())
	// 	// 	if err != nil {
	// 	// 		log.Println("err:", err)
	// 	// 	}
	// 	// }
	// }
	//
	// return 0, nil
}
