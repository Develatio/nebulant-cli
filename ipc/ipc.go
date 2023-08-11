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

package ipc

import (
	"fmt"
	"io"
	"net"
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

type IPCConsumer struct {
	ID     string
	Stream chan *PipeData
}

type IPC struct {
	uuid      string
	consumers map[string]*IPCConsumer
	l         net.Listener
	Errors    chan error
}

func (p *IPC) NewConsumer(ipcc *IPCConsumer) {
	p.consumers[ipcc.ID] = ipcc
}

func (p *IPC) OutConsumer(ipcc *IPCConsumer) {
	delete(p.consumers, ipcc.ID)
}

func (p *IPC) GetUUID() string {
	return p.uuid
}

func (p *IPC) Close() error {
	return p.l.Close()
}

func (p *IPC) Accept() error {
	defer p.l.Close()
	for {
		con, err := p.l.Accept()
		if err != nil {
			p.Errors <- err
		}
		go p.serve(con)
	}
}

func (p *IPC) serve(con net.Conn) {
	defer con.Close()
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

func NewIPCServer(euuid string) (*IPC, error) {
	ipc := &IPC{
		uuid:      euuid,
		consumers: make(map[string]*IPCConsumer),
		Errors:    make(chan error),
	}
	l, err := ipc.Listen()
	if err != nil {
		return nil, err
	}
	ipc.l = l
	return ipc, nil
}
