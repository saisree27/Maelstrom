package engine

import "fmt"

func printBitBoard(u u64) {
	s := "\n"
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			// lookup bit at square i + j
			if u&u64(1<<(i+j)) != 0 {
				s += "1" + " "
			} else {
				s += "0" + " "
			}
		}
		s += "\n"
	}
	fmt.Print(s)
}
