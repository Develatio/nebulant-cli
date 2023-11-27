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
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

var MAXWRITESIZE = 1024
var MAXREADSIZE = 1024

type ioWSrw struct {
	conn *websocket.Conn
}

func (i *ioWSrw) Write(p []byte) (int, error) {
	err := i.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (i *ioWSrw) Read(p []byte) (n int, err error) {
	_, m, err := i.conn.ReadMessage()
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

func parseDebuggFs() (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("debugterm", flag.ExitOnError)
	err := fs.Parse(flag.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func DebuggerCmd() (int, error) {

	u := url.URL{Scheme: "ws", Host: "localhost:6565", Path: ""}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	c.SetReadLimit(int64(MAXREADSIZE))
	if err != nil {
		return 1, err
	}
	defer c.Close()

	iows := &ioWSrw{
		conn: c,
	}

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
	fmt.Println("looping")
	// raw term
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)
	os.Setenv("TERM", "vt100")

	go func() {
		// var bff []byte
		// var buff = bytes.NewBuffer(bff)
		reader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := reader.ReadRune()
			if err != nil {
				fmt.Println("ERRRRRR 1")
				return
			}
			// buff.WriteRune(char)
			// fmt.Println("writing to iows", char)
			// fmt.Printf("Rune read: %d", char)
			// fmt.Println("----")
			iows.Write([]byte(string(char)))
		}

	}()

	for {
		// bff := make([]byte, MAXREADSIZE)
		// _, err := iows.Read(bff)
		// if err != nil {
		// 	return 1, err
		// }
		// fmt.Println(bff)
		// fmt.Printf("v:%v d:%d s:%s", bff, bff, bff)
		// fmt.Printf("%s", bff)
		io.Copy(os.Stdout, iows)
		// switch buff.String() {
		// case "exit":
		// 	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// 	if err != nil {
		// 		log.Println("write close:", err)
		// 		return 1, err
		// 	}
		// 	select {
		// 	case <-done:
		// 	case <-time.After(time.Second):
		// 	}
		// 	break L
		// default:
		// 	err := c.WriteMessage(websocket.BinaryMessage, buff.Bytes())
		// 	if err != nil {
		// 		log.Println("err:", err)
		// 	}
		// }
	}
	//
	// return 0, nil
}
