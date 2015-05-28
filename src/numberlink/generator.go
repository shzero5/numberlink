package main

import "fmt"
import "os"
import "strconv"
import "math/rand"
import "time"
import "strings"

var (
	SIGMA = [92]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '!', '"', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '/', ':', ';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~'}
	DX    = [4]int{0, 1, 0, -1}
	DY    = [4]int{-1, 0, 1, 0}
)

func square(x int) int {
	return x * x
}

// Generate and prints a width x height puzzle of numberlink
// The algorithm works as follows:
// 1) First the board is tiled with 2x1 dominos in a simple, deterministic way.
//    If this is not possible (on an odd area paper), the bottom right corner
//    is left unconnected
// 2) Then the dominos are randomly shuffled by flipping random pairs of
//    neighbours. This is (obviously) not done in the case of width or height
//    equal to 1
// 3) Now, in the case of an odd area paper, the bottom right corner is
//    attached to one of its neighbour dominos. This will always be possible.
// 4) Finally we can start finding random paths through the dominos, combining
//    them as we pass through. Special care is taken not to connect 'touching
//    flows' which would create puzzles that 'double back on themselves'
// 5) Before the puzzle is printed we 'compact' the range of colors used, as
//    much as possible
// 6) The puzzle is printed by replacing all positions that aren't flow-heads
//    with a .
func Generate(width, height int) [][]int {
	if width == 0 || height == 0 || width == 1 && height == 1 {
		return nil
	}
	rand.Seed(time.Now().UTC().UnixNano())
	table := tile(width, height)
	shuffle(table)
	oddCorner(table)
	oddDomino(table)
	findFlows(table)
	return table
}

func print(table [][]int) ([]string, []string, error) {
	width, height := len(table[0]), len(table)
	colorsUsed := flatten(table)
	if colorsUsed > len(SIGMA) {
		return nil, nil, fmt.Errorf("Error: Not enough printable characters to print puzzle")
	}

	pzzl := make([]string, height)
	sltn := make([]string, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sltn[y] = sltn[y] + string(SIGMA[table[y][x]])
			if isFlowHead(x, y, table) {
				pzzl[y] = pzzl[y] + string(SIGMA[table[y][x]])
			} else {
				pzzl[y] = pzzl[y] + "."
			}
		}
	}

	return pzzl, sltn, nil
}

func tile(width, height int) [][]int {
	table := make([][]int, height)
	for y := 0; y < height; y++ {
		table[y] = make([]int, width)
	}
	// Start with simple vertical tiling
	alpha := int(0)
	for y := 0; y < height-1; y += 2 {
		for x := 0; x < width; x++ {
			table[y][x] = alpha
			table[y+1][x] = alpha
			alpha += 1
		}
	}
	// Add padding in case of odd height
	if height%2 == 1 {
		for x := 0; x < width-1; x += 2 {
			table[height-1][x] = alpha
			table[height-1][x+1] = alpha
			alpha += 1
		}
		// In case of odd width, add a single in the corner.
		// We will merge it into a real flow after shuffeling
		if width%2 == 1 {
			table[height-1][width-1] = alpha
		}
	}
	return table
}

func shuffle(table [][]int) {
	width, height := len(table[0]), len(table)
	if width == 1 || height == 1 {
		return
	}
	for i := 0; i < square(width*height); i++ {
		x, y := rand.Intn(width-1), rand.Intn(height-1)
		if table[y][x] == table[y][x+1] && table[y+1][x] == table[y+1][x+1] {
			// Horizontal case
			// aa \ ab
			// bb / ab
			table[y+1][x] = table[y][x]
			table[y][x+1] = table[y+1][x+1]
		} else if table[y][x] == table[y+1][x] && table[y][x+1] == table[y+1][x+1] {
			// Vertical case
			// ab \ aa
			// ab / bb
			table[y][x+1] = table[y][x]
			table[y+1][x] = table[y+1][x+1]
		}
	}
}

func oddCorner(table [][]int) {
	width, height := len(table[0]), len(table)
	if width%2 == 1 && height%2 == 1 {
		// Horizontal case:
		// aax
		if width > 2 && table[height-1][width-3] == table[height-1][width-2] {
			table[height-1][width-1] = table[height-1][width-2]
		}
		// Vertical case:
		// ab
		// ax
		if height > 2 && table[height-3][width-1] == table[height-2][width-1] {
			table[height-1][width-1] = table[height-2][width-1]
		}
	}
}

func oddDomino(table [][]int) {
	width, height := len(table[0]), len(table)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if inside(x+2, y+1, width, height) {
				if table[y][x] == table[y+1][x] && table[y][x+1] == table[y+1][x+1] && table[y][x+2] == table[y+1][x+2] && table[y][x] != table[y][x+1] && table[y][x+1] != table[y][x+2] {
					if rand.Intn(1) == 0 {
						table[y][x+1] = table[y][x]
						table[y][x+2] = table[y][x]
						table[y+1][x] = table[y+1][x+2]
						table[y+1][x+1] = table[y+1][x+2]
					} else {
						table[y][x+1] = table[y][x]
						table[y+1][x+1] = table[y][x+2]
					}
				}
			}
			/*if inside(x+1, y+2, width, height) {
				if table[y][x] == table[y][x+1] && table[y+1][x] == table[y+1][x+1] && table[y+2][x] == table[y+2][x+1] && table[y][x] != table[y+1][x] && table[y+1][x] != table[y+2][x] {
					table[y+1][x] = table[y][x]
					table[y+1][x+1] = table[y+2][x]
				}
			}*/
		}
	}
}

func findFlows(table [][]int) {
	width, height := len(table[0]), len(table)
	//for i := 0; i < 10; i++ {
	for _, p := range rand.Perm(width * height) {
		x, y := p%width, p/width
		if isFlowHead(x, y, table) {
			layFlow(x, y, table)
		}
	}
	//}
}

// 基点が否かを判定
func isFlowHead(x, y int, table [][]int) bool {
	width, height := len(table[0]), len(table)
	degree := 0
	for i := 0; i < 4; i++ {
		x1, y1 := x+DX[i], y+DY[i]
		if inside(x1, y1, width, height) && table[y1][x1] == table[y][x] {
			degree += 1
		}
	}
	return degree < 2
}

func inside(x, y int, width, height int) bool {
	return 0 <= x && x < width && 0 <= y && y < height
}

func layFlow(x, y int, table [][]int) {
	width, height := len(table[0]), len(table)
	for _, i := range rand.Perm(4) {
		x1, y1 := x+DX[i], y+DY[i]
		if inside(x1, y1, width, height) && canConnect(x, y, x1, y1, table) {
			fill(x1, y1, table[y][x], table)
			x2, y2 := follow(x1, y1, x, y, table)
			layFlow(x2, y2, table)
			return
		}
	}
}

func canConnect(x1, y1, x2, y2 int, table [][]int) bool {
	width, height := len(table[0]), len(table)
	// Check (x1,y2) and (x2,y2) are flow heads
	if table[y1][x1] == table[y2][x2] {
		return false
	}
	if !isFlowHead(x1, y1, table) || !isFlowHead(x2, y2, table) {
		return false
	}
	for y3 := 0; y3 < height; y3++ {
		for x3 := 0; x3 < width; x3++ {
			for i := 0; i < 4; i++ {
				x4, y4 := x3+DX[i], y3+DY[i]
				if inside(x4, y4, width, height) &&
					!(x3 == x1 && y3 == y1 && x4 == x2 && y4 == y2) &&
					table[y3][x3] == table[y1][x1] && table[y4][x4] == table[y2][x2] {
					return false
				}
			}
		}
	}
	return true
}

func fill(x, y int, alpha int, table [][]int) {
	width, height := len(table[0]), len(table)
	orig := table[y][x]
	table[y][x] = alpha
	for i := 0; i < 4; i++ {
		x1, y1 := x+DX[i], y+DY[i]
		if inside(x1, y1, width, height) && table[y1][x1] == orig {
			fill(x1, y1, alpha, table)
		}
	}
}

func follow(x, y, x0, y0 int, table [][]int) (int, int) {
	width, height := len(table[0]), len(table)
	for i := 0; i < 4; i++ {
		x1, y1 := x+DX[i], y+DY[i]
		if inside(x1, y1, width, height) && !(x1 == x0 && y1 == y0) &&
			table[y][x] == table[y1][x1] {
			return follow(x1, y1, x, y, table)
		}
	}
	return x, y
}

// Expects the table to be valid as generated by the above functions, in the following way:
// * Areas with the same value must be grouped in such a way that you can change them with `fill`
// * Values must be in the range [0...)
func flatten(table [][]int) int {
	width, height := len(table[0]), len(table)
	// Flatten all the flows at -iota-1 so we don't
	// accidentially merge something
	alpha := -1
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if table[y][x] >= 0 {
				fill(x, y, alpha, table)
				alpha -= 1
			}
		}
	}
	// Then invert to get what we actually wanted
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			table[y][x] = -table[y][x] - 1
		}
	}
	return -alpha - 1
}

func member(n int, xs []int) bool {
    for _, x := range xs {
        if n == x { return true }
    }
    return false
}

func removeDup(xs []int) []int {
    ys := make([]int, 0, len(xs))
    for _, x := range xs {
        if !member(x, ys) {
            ys = append(ys, x)
        }
    }
    return ys
}

func isValid(pzzl [][]int) bool {
	width, height := len(pzzl[0]), len(pzzl)
	for y0 := 0; y0 < height; y0++ {
		var nums []int
		for x0 := 0; x0 < width; x0++ {
			if isFlowHead(x0, y0, pzzl) {
				nums = append(nums, pzzl[y0][x0])
			}
			for i := 0; i < 4; i++ {
				x1, y1 := x0+DX[i], y0+DY[i]
				if inside(x1, y1, width, height) && !(x1 == x0 && y1 == y0) && pzzl[y0][x0] == pzzl[y1][x1] {
					if isFlowHead(x0, y0, pzzl) && isFlowHead(x1, y1, pzzl) {
						return false
					}
				}
			}
		}
		subnums := removeDup(nums)
		if len(nums) != len(subnums) {
			return false
		}
	}
	return true
}

func isValidNumber(pzzl []string) bool {
	maxNum := 0
	for _, line := range pzzl {
		chars := strings.Split(line, "")
		for _, char := range chars {
			if char != "." {
				i, _ := strconv.Atoi(char)
				if i >= 9 {
					return false
				}
				if i > maxNum {
					maxNum = i
				}
			}
		}
	}
	return maxNum >= 2
}

func displayPuzzle(pzzl []string) {
	fmt.Print("[")
	flag1 := true
	for _, line := range pzzl {
		if flag1 {
			flag1 = false
		} else {
			fmt.Print(",")
		}
		fmt.Print("[")
		chars := strings.Split(line, "")
		flag2 := true
		for _, char := range chars {
			if flag2 {
				flag2 = false
			} else {
				fmt.Print(",")
			}
			if char == "." {
				fmt.Print("0")
			} else {
				i, _ := strconv.Atoi(char)
				fmt.Print(i+1)
			}
		}
		fmt.Print("]")
	}
	fmt.Println("],")
}

func gen(size int) {
	table := Generate(size, size)
	pzzl, _, _ := print(table)

	for !isValid(table) || !isValidNumber(pzzl) {
		table = Generate(size, size)
		pzzl, _, _ = print(table)
	}

	displayPuzzle(pzzl)

	/*if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Println(len(pzzl[0]), len(pzzl))
		for _, line := range pzzl {
			fmt.Println(line)
		}
	}*/
	return
}


func main() {
	size, _ := strconv.Atoi(os.Args[1])
	for i := 0; i < 600; i++ {
		gen(size)
	}
}
