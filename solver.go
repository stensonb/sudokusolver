package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const Unk uint8 = 0

type SudokuError interface {
	Error() string
	Code() int
}

type InvalidBoard string
type CannotSolveBoardError string

func (i InvalidBoard) Error() string {
	return "invalid board"
}

func (i InvalidBoard) Code() int {
	return 10
}

func (c CannotSolveBoardError) Error() string {
	return "cannot solve board"
}

func (i CannotSolveBoardError) Code() int {
	return 11
}

type Board struct {
	vals [][]uint8
}

func (b *Board) String() string {
	ans := ""
	for i := range b.vals {
		for j := range b.vals[i] {
			ans += fmt.Sprint(b.vals[i][j], " ")
		}
		ans += "\n"
	}
	return ans
}

// returns true if all values are filled in
// returns false if ANY value is Unk
func (b *Board) Solved() bool {
	for i := range b.vals {
		for j := range b.vals[i] {
			if b.vals[i][j] == Unk {
				return false
			}
		}
	}
	return true
}

// returns true if all values honor the rules of the game
// returns false if ANY value is invalid
func (b *Board) Valid() bool {
	return b.validRows() && b.validCols() && b.validPods()
}

func (b *Board) validRows() bool {
	// if any individual row is invalid, return false
	for _, v := range b.vals {
		if !validSet(v) {
			return false
		}
	}

	// otherwise, it's valid
	return true
}

// we assume b.vals is a array where height == width
func (b *Board) validCols() bool {
	// build array of sets representing columns
	cols := make([][]uint8, len(b.vals))
	for s := range cols {
		cols[s] = []uint8{}
	}

	// fill in arrays of sets with proper values
	for r := range b.vals {
		for c := range b.vals[r] {
			cols[c] = append(cols[c], b.vals[r][c])
		}
	}

	// if !validSet() on any set, return false
	for c := range cols {
		if !validSet(cols[c]) {
			return false
		}
	}

	// otherwise, all columns are valid
	return true
}

func (b *Board) validPods() bool {
	// build array of sets representing pods
	pods := make([][]uint8, len(b.vals))
	for s := range pods {
		pods[s] = []uint8{}
	}

	// fill in arrays of sets with proper values
	for r := range b.vals {
		for c := range b.vals[r] {
			// some function which maps r,c to pod index
			pods[whichPod(r, c)] = append(pods[whichPod(r, c)], b.vals[r][c])
		}
	}

	// if !validSet() on any set, return false
	for p := range pods {
		if !validSet(pods[p]) {
			return false
		}
	}

	// otherwise, all pods are valid
	return true
}

// recursive approach to solving this board
func SolveIt(b Board) (Board, SudokuError) {
	s := b.Solved()
	v := b.Valid()

	switch {
	case s && v: // base case
		return b, nil
	case !v: // invalid board
		return Board{}, new(InvalidBoard)
	default: // not solved, but still valid...recurse!
		// create a copy to modify
		bnew := Copy(b)
		// find an Unk
		for i := range b.vals {
			for j := range b.vals[i] {
				if bnew.vals[i][j] == Unk {
					// try values, starting at 1, going up to len(b.vals)
					for k := uint8(1); k <= uint8(len(bnew.vals)); k++ {
						bnew.vals[i][j] = k
						bn, e := SolveIt(bnew)
						if e == nil && bn.Solved() && bn.Valid() {
							return bn, nil
						}
					}
				}
			}
		}
	}

	// every combination has been tried, cannot solve this board
	return Board{}, new(CannotSolveBoardError)
}

func Copy(b Board) Board {
	ans := new(Board)
	ans.vals = make([][]uint8, len(b.vals))

	for i := range ans.vals {
		ans.vals[i] = make([]uint8, len(b.vals))
		for j := range ans.vals[i] {
			ans.vals[i][j] = b.vals[i][j]
		}
	}
	return *ans
}

// returns an pod index based on r,c
func whichPod(r, c int) int {
	switch {
	case r < 3:
		switch {
		case c < 3:
			return 0
		case c > 5:
			return 2
		default: // middle column
			return 1
		}
	case r > 5:
		switch {
		case c < 3:
			return 6
		case c > 5:
			return 8
		default: // middle column
			return 7
		}
	default: // middle row
		switch {
		case c < 3:
			return 3
		case c > 5:
			return 5
		default: // middle column
			return 4
		}
	}
}

// returns true if the array passed in contains zero duplicates
// returns false if ANY element in the array is a duplicate (except Unk)
func validSet(r []uint8) bool {
	vals := make(map[uint8]bool)

	for _, v := range r {
		// if we've seen this value already, then the set is invalid
		// ie, cannot have two 1s in the same set
		if vals[v] && v != Unk {
			return false
		}

		// set it true to ensure only one in this set
		vals[v] = true
	}

	return true
}

// returns a valid sudoku board
// returns an error if user input fails to parse to a valid board
// ANY rune which is not a positive integer will be set interpreted as Unk
func GetBoardFromInput() (*Board, SudokuError) {
	ans := new(Board)
	ans.vals = [][]uint8{}

	reader := bufio.NewReader(os.Stdin)

	line_width := -1
	input_lines := 0
	for line_width != input_lines {
		line, _ := reader.ReadString('\n')

		// remove leading and trailing spaces/newlines
		digits := strings.Split(strings.TrimSpace(line), " ")

		if line_width == -1 { // first time through, set width
			line_width = len(digits)
		}

		// subsequent inputs lines must equal the length as the first line
		// if not, return InvalidBoard
		if len(digits) != line_width {
			return &Board{}, new(InvalidBoard)
		}

		var digit int
		var err error

		ints := []uint8{}
		for i := range digits {
			// attempt to convert to integers
			digit, err = strconv.Atoi(digits[i])

			if err != nil {
				// it's not an integer, so make it Unk
				digit = int(Unk)
			}

			// cannot be a negative number
			if digit < 0 {
				digit = int(Unk)
			}

			ints = append(ints, uint8(digit))
		}

		ans.vals = append(ans.vals, ints)
		input_lines += 1
	}

	return ans, nil
}

func main() {
	b, e := GetBoardFromInput()
	if e != nil {
		fmt.Println(e)
		os.Exit(e.Code())
	}

	solvedBoard, e := SolveIt(*b)
	if e != nil {
		fmt.Println(e)
		os.Exit(e.Code())
	}

	fmt.Print(solvedBoard.String())
}
