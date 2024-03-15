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
