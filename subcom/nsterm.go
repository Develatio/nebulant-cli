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
