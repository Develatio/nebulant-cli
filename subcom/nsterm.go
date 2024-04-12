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

// func NSTerm(nblc *subsystem.NBLcommand) (int, error) {
// 	// raw term, pty will be emulated
// 	oldState, err := term.MakeRaw(int(nebulant_term.GenuineOsStdin.Fd()))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer term.Restore(int(nebulant_term.GenuineOsStdin.Fd()), oldState)

// 	vpty := nsterm.NewVirtPTY()
// 	mfd := vpty.MustarFD()
// 	go func() {
// 		// fmt.Println("stdin on")
// 		io.Copy(mfd, os.Stdin)
// 		// fmt.Println("stdin off")
// 	}()
// 	go func() {
// 		// fmt.Println("stdout on")
// 		io.Copy(os.Stdout, mfd)
// 		// fmt.Println("stdout off")
// 	}()
// 	// WIP, TODO: send this ldisc to
// 	// shell or leave shell to set it
// 	// to allow shell handle line buffer
// 	// TODO: bring defaultldisc hability
// 	// to work with line buff
// 	// ldisc := &DefaultLdisc{}
// 	// vpty.SetLDisc(ldisc)
// 	// vpty.OpenMustar(nebulant_term.GenuineOsStdin, nebulant_term.GenuineOsStdout)
// 	// vstdin, vstdout := vpty.OpenSluva(false)

// 	sfd := vpty.SluvaFD()

// 	defer mfd.Close()
// 	defer sfd.Close()

// 	// fmt.Println("running shell")
// 	return nsterm.NSShell(vpty, sfd, sfd)
// }
