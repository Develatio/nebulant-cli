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

package nsterm

// I+D+I
var HandleScapeMode = false

// func readESCstdin(mfd *PortFD) {
// 	reader := bufio.NewReader(os.Stdin)
// 	_c := make(chan rune, 10)
// 	_e := make(chan error, 10)

// 	go func() {
// 		for {
// 			ccc, _, eee := reader.ReadRune()
// 			if eee != nil {
// 				_e <- eee
// 				continue
// 			}
// 			_c <- ccc
// 		}
// 	}()

// 	for {
// 		select {
// 		case char := <-_c:
// 			// ESC sequence
// 			if char == 27 {
// 				// collect
// 				time.Sleep(512 * time.Microsecond)
// 				esq_seq_size := len(_c)
// 				if esq_seq_size > 0 {
// 					esc_seq := []rune{27}
// 					for i := 0; i < esq_seq_size; i++ {
// 						esc_seq = append(esc_seq, <-_c)
// 					}
// 					esc_seq = append(esc_seq, []rune("\n")...)
// 					// write back entire esc secuence
// 					mfd.Write([]byte(string(esc_seq)))
// 					continue
// 				}
// 				// write esc char
// 				mfd.Write([]byte(string(char)))
// 				continue
// 			}
// 		case _err := <-_e:
// 			mfd.Write([]byte(errors.Join(fmt.Errorf("term read err"), _err).Error()))
// 		default:
// 			<-time.After(100 * time.Microsecond)
// 		}
// 	}
// }

// func NSTerm(nblc *subsystem.NBLcommand) (int, error) {
// 	// raw term, pty will be emulated
// 	oldState, err := term.MakeRaw(int(nebulant_term.GenuineOsStdin.Fd()))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer term.Restore(int(nebulant_term.GenuineOsStdin.Fd()), oldState)

// 	vpty := NewVirtPTY()
// 	mfd := vpty.MustarFD()

// 	if HandleScapeMode {
// 		go func() {
// 			io.Copy(mfd, os.Stdin)
// 		}()
// 	} else {
// 		// capture escape sequences
// 		go readESCstdin(mfd)
// 	}
// 	go func() {
// 		io.Copy(os.Stdout, mfd)
// 	}()

// 	sfd := vpty.SluvaFD()

// 	defer mfd.Close()
// 	defer sfd.Close()

// 	return NSShell(vpty, sfd, sfd)
// }
