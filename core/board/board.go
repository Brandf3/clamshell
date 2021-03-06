package board

import (
	"container/list"
	"fmt"
	"strings"

	"github.com/otrego/clamshell/core/color"
	"github.com/otrego/clamshell/core/move"
	"github.com/otrego/clamshell/core/point"
)

// Board Contains the board, capturesStones, and ko
// ko contains a point that is illegal to recapture due to Ko.
type Board struct {
	board [][]color.Color
	ko    *point.Point
}

// NewBoard creates a new size x size board.
func NewBoard(size int) *Board {
	board := Board{
		make([][]color.Color, size),
		nil,
	}

	for i := 0; i < size; i++ {
		board.board[i] = make([]color.Color, size)
	}
	return &board
}

// PlaceStone adds a stone to the board
// and removes captured stones (if any).
// returns the captured stones, or err
// if any Go (baduk) rules were broken
func (b *Board) PlaceStone(m *move.Move) ([]*point.Point, error) {
	var ko *point.Point = b.ko
	b.ko = nil

	if !b.inBounds(m.Point()) {
		return nil, fmt.Errorf("move %v out of bounds for %dx%d board",
			m.Point(), len(b.board[0]), len(b.board))
	}
	if b.colorAt(m.Point()) != color.Empty {
		return nil, fmt.Errorf("move %v already occupied", m.Point())
	}

	b.setColor(m)
	capturedStones := b.findCapturedGroups(m)
	if len(capturedStones) == 0 && len(b.capturedStones(m.Point())) != 0 {
		b.setColor(move.NewMove(color.Empty, m.Point()))
		return nil, fmt.Errorf("move %v is suicidal", m.Point())
	}
	if len(capturedStones) == 1 {
		b.ko = m.Point()
		if ko != nil && *ko == *(capturedStones[0]) {
			b.setColor(move.NewMove(color.Empty, m.Point()))
			return nil, fmt.Errorf("%v is an illegal ko move", m.Point())
		}
	}

	b.removeCapturedStones(capturedStones)
	return capturedStones, nil
}

// findCapturedGroups returns the groups captured by *Move m.
func (b *Board) findCapturedGroups(m *move.Move) []*point.Point {
	pt := m.Point()

	points := b.getNeighbors(pt)
	capturedStones := make([]*point.Point, 0)
	for _, point := range points {
		if b.inBounds(point) {
			capturedStones = append(capturedStones, b.capturedStones(point)...)
		}
	}
	return capturedStones
}

// removeCapturedStones removes the captured stones from
// the board.
func (b *Board) removeCapturedStones(capturedStones []*point.Point) {
	for _, point := range capturedStones {
		b.setColor(move.NewMove(color.Empty, point))
	}
}

// CapturedStones returns the captured stones in group containing Point pt.
// returns nil if no stones were captured.
func (b *Board) capturedStones(pt *point.Point) []*point.Point {
	expanded := make(map[point.Point]bool)

	// current group color
	c := b.colorAt(pt)

	queue := list.New()
	queue.PushBack(pt)
	for queue.Len() > 0 {
		e := queue.Front()
		queue.Remove(e)
		pt1, ok := e.Value.(*point.Point)
		if !ok {
			panic("e.Value was not of type point.Point")
		}

		if !b.inBounds(pt1) {
			continue
		} else if b.colorAt(pt1) == color.Empty {
			// Liberty has been found, no need to continue search
			return nil
		} else if b.colorAt(pt1) == c && !expanded[*pt1] {
			expanded[*pt1] = true
			points := b.getNeighbors(pt1)
			for _, point := range points {
				queue.PushBack(point)
			}
		}
	}

	// The stones that were captured
	stoneGroup := make([]*point.Point, len(expanded))
	i := 0
	for key := range expanded {
		stoneGroup[i] = point.New(key.X(), key.Y())
		i++
	}
	return stoneGroup
}

// inBounds returns true if x and y are in bounds
// on the board, false otherwise.
func (b *Board) inBounds(pt *point.Point) bool {
	var x, y int = int(pt.X()), int(pt.Y())
	return x < len(b.board[0]) && y < len(b.board) &&
		x >= 0 && y >= 0
}

// colorAt returns the color at point pt.
func (b *Board) colorAt(pt *point.Point) color.Color {
	var x, y int = int(pt.X()), int(pt.Y())
	return b.board[y][x]
}

// setColor sets the color m.Color at point m.Point.
func (b *Board) setColor(m *move.Move) {
	var x, y int = int(m.Point().X()), int(m.Point().Y())
	b.board[y][x] = m.Color()
}

// getNeighbors returns a list of points neighboring point pt.
// Neighboring points could be out of bounds.
func (b *Board) getNeighbors(pt *point.Point) []*point.Point {
	points := make([]*point.Point, 4)
	points[0] = point.New(pt.X()+1, pt.Y())
	points[1] = point.New(pt.X()-1, pt.Y())
	points[2] = point.New(pt.X(), pt.Y()+1)
	points[3] = point.New(pt.X(), pt.Y()-1)

	return points
}

// GetFullBoardState returns an array of all the current stone positions.
func (b *Board) GetFullBoardState() []*move.Move {
	moves := make([]*move.Move, 0)

	for i := 0; i < len(b.board); i++ {
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] != color.Empty {
				moves = append(moves,
					move.NewMove(b.board[i][j], point.New(int64(j), int64(i))))
			}
		}
	}

	return moves
}

// String returns a string representation of this board.
// For example:
//
//    b.Board {{B, W, B,  },
//             {W,  , B, B},
//             { ,  , W,  },
//             {B,  , W,  }}
//
//    Becomes  [B W B .]
//             [W . B B]
//             [. . W .]
//             [B . W .]
func (b *Board) String() string {
	var sb strings.Builder
	for i := 0; i < len(b.board); i++ {
		// To increase useability of this String function,
		// color.Empty is converted from "" to ".".
		str := make([]string, len(b.board[0]))
		for j := 0; j < len(b.board[0]); j++ {
			if b.board[i][j] == color.Empty {
				str[j] = "."
			} else {
				str[j] = string(b.board[i][j])
			}
		}
		sb.WriteString(fmt.Sprintf("%v\n", str))
	}
	return strings.TrimSpace(sb.String())
}
