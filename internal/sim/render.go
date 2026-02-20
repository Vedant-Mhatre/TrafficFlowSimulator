package sim

import (
	"fmt"
	"strings"
)

func RenderGrid(cfg Config, vehicles []Vehicle, light TrafficLight, step int) {
	width := cfg.Grid.Width
	height := cfg.Grid.Height

	grid := make([][]rune, height)
	for y := range grid {
		grid[y] = make([]rune, width)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}

	ix := width / 2
	iy := height / 2

	for y := 0; y < height; y++ {
		grid[y][ix] = '|'
	}
	for x := 0; x < width; x++ {
		grid[iy][x] = '-'
	}
	if light.VerticalGreen {
		grid[iy][ix] = 'G'
	} else {
		grid[iy][ix] = 'R'
	}

	for i := range vehicles {
		v := vehicles[i]
		if v.X >= 0 && v.X < width && v.Y >= 0 && v.Y < height {
			grid[v.Y][v.X] = directionRune(v.Direction)
		}
	}

	fmt.Print("\033[H\033[2J")
	fmt.Printf("Step %d/%d (vertical=%v)\n", step+1, cfg.Steps, light.VerticalGreen)
	for y := 0; y < height; y++ {
		var row strings.Builder
		for x := 0; x < width; x++ {
			row.WriteRune(grid[y][x])
		}
		fmt.Println(row.String())
	}
}

func directionRune(d Direction) rune {
	switch d {
	case Up:
		return '^'
	case Down:
		return 'v'
	case Left:
		return '<'
	case Right:
		return '>'
	default:
		return 'V'
	}
}
