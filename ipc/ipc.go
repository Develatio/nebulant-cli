// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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

package ipc

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/cast"
)

type PipeData struct {
	IPCSID  string
	IPCCID  string
	COMMAND string
	VARNAME string
	c       net.Conn
}

func (d *PipeData) Resp(r string) error {
	_, err := d.c.Write([]byte(r))
	if err != nil {
		return err
	}
	return nil
}

func (d *PipeData) Close() error {
	return d.c.Close()
}

func (d *PipeData) RespClose(r string) error {
	err := d.Resp(r)
	if err != nil {
		return err
	}
	return d.Close()
}

type IPCConsumer struct {
	ID     string
	Stream chan *PipeData
}

func (c *IPCConsumer) ExposeStoreVars(store base.IStore) chan bool {
	out := make(chan bool)
	go func() {
	L:
		for { // Infine loop until break L
			select { // Loop until a case ocurrs.
			case data := <-c.Stream:
				if data.COMMAND == "readvar" {
					resp := "{{ " + data.VARNAME + " }}"
					err := store.Interpolate(&resp)
					if err != nil {
						resp = "\x10"
					}
					if resp == "{{ "+data.VARNAME+" }}" || resp == "" {
						resp = "\x10"
					}
					err = data.RespClose(resp)
					if err != nil {
						if err != io.EOF {
							break L
						}
					}
				}
			case <-out:
				break L
			default:
				time.Sleep(200000 * time.Microsecond)
			}
		}
	}()
	return out
}

type IPC struct {
	uuid      string
	consumers map[string]*IPCConsumer
	l         net.Listener
	Errors    chan error
	closed    bool
}

func (p *IPC) IsClosed() bool {
	return p.closed
}

func (p *IPC) SetListener(l net.Listener) {
	p.l = l
}

func (p *IPC) AppendConsumer(ipcc *IPCConsumer) {
	p.consumers[ipcc.ID] = ipcc
}

func (p *IPC) OutConsumer(ipcc *IPCConsumer) {
	delete(p.consumers, ipcc.ID)
}

func (p *IPC) GetUUID() string {
	return p.uuid
}

func (p *IPC) Close() error {
	if p.l == nil {
		return nil
	}
	err := p.l.Close()
	p.l = nil
	p.closed = true
	return err
}

func (p *IPC) Accept() error {
	p.closed = false
	defer func() {
		err := p.Close()
		if err != nil {
			p.Errors <- err
		}
	}()
	for {
		if p.l == nil {
			break
		}
		con, err := p.l.Accept()
		if err != nil {
			p.Errors <- err
			continue
		}
		go p.serve(con)
	}
	return nil
}

func (p *IPC) serve(con net.Conn) {
	defer func() {
		if con != nil {
			err := con.Close()
			if err != nil {
				cast.LogWarn(errors.Join(fmt.Errorf("ipc server connection close err"), err).Error(), nil)
			}
		}
	}()
	buf := make([]byte, 512)
	// var buff bytes.Buffer
	for {
		n, err := con.Read(buf)
		if err != nil {
			if err != io.EOF {
				p.Errors <- err
			}
			break
		}
		str := string(buf[:n])
		ppd := &PipeData{c: con}
		_, err = fmt.Sscanf(str, "%s %s %s %s", &ppd.IPCSID, &ppd.IPCCID, &ppd.COMMAND, &ppd.VARNAME)
		if err != nil {
			p.Errors <- err
			continue
		}
		if _, exists := p.consumers[ppd.IPCCID]; exists {
			p.consumers[ppd.IPCCID].Stream <- ppd
		} else {
			err := ppd.Resp("")
			if err != nil {
				p.Errors <- err
			}
		}
	}
}

func NewListenerIPCServer(l net.Listener, id string) (*IPC, error) {
	var err error
	ipc := &IPC{
		uuid:      id,
		consumers: make(map[string]*IPCConsumer),
		Errors:    make(chan error),
	}
	if l == nil {
		l, err = ipc.listen()
		if err != nil {
			return nil, err
		}
	}
	ipc.l = l
	return ipc, nil
}

func NewIPCServer() (*IPC, error) {
	return NewListenerIPCServer(nil, fmt.Sprintf("%d", rand.Int())) // #nosec G404 -- Weak random is OK here
}
