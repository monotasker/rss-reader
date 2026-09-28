package tui

import "os"

func ReadKey() string {
	buf := make([]byte, 3)
	os.Stdin.Read(buf)

	if buf[0] == 27 { // escape sequence
		switch string(buf) {
		case "\033[A":
			return "up"
		case "\033[B":
			return "down"
		default:
			return "escape"
		}
	}
	return string(buf)
}
