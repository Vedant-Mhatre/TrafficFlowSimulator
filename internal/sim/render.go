package sim

import (
	"fmt"
	"sort"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

type RenderStats struct {
	ScenarioName         string
	Step                 int
	TotalSteps           int
	VerticalGreen        bool
	SpawnedVehicles      int
	CompletedVehicles    int
	ActiveVehicles       int
	BlockedBySignal      int
	BlockedByTraffic     int
	PotentialCollisions  int
	MaxQueueOverall      int
	AverageNetworkSpeed  float64
	ThroughputPer100Step float64
	LaneQueue            map[Direction]int
	LaneActive           map[Direction]int
}

func RenderGrid(cfg Config, vehicles []Vehicle, light TrafficLight, stats RenderStats) {
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
	printHeader(stats)
	printRoadFrame(grid, light)
	printFooter(stats)
}

func printHeader(stats RenderStats) {
	phase := colorRed + "HORIZONTAL GREEN" + colorReset
	if stats.VerticalGreen {
		phase = colorGreen + "VERTICAL GREEN" + colorReset
	}
	fmt.Printf("%s%sTrafficFlowSimulator Terminal Dashboard%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%sScenario:%s %s | %sStep:%s %d/%d | %sPhase:%s %s\n",
		colorBold, colorReset, stats.ScenarioName,
		colorBold, colorReset, stats.Step, stats.TotalSteps,
		colorBold, colorReset, phase,
	)
	fmt.Printf("%sVehicles%s spawned=%d completed=%d active=%d | %sSpeed%s avg=%.3f | %sThroughput%s %.2f/100\n",
		colorBold, colorReset, stats.SpawnedVehicles, stats.CompletedVehicles, stats.ActiveVehicles,
		colorBold, colorReset, stats.AverageNetworkSpeed,
		colorBold, colorReset, stats.ThroughputPer100Step,
	)
	fmt.Printf("%sBlockers%s signal=%d traffic=%d | %sConflicts%s potential=%d | %sMax Queue%s %d\n",
		colorBold, colorReset, stats.BlockedBySignal, stats.BlockedByTraffic,
		colorBold, colorReset, stats.PotentialCollisions,
		colorBold, colorReset, stats.MaxQueueOverall,
	)
	fmt.Println()
}

func printRoadFrame(grid [][]rune, light TrafficLight) {
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])
	hLine := strings.Repeat("-", width)
	fmt.Printf("%s+%s+%s\n", colorDim, hLine, colorReset)
	for y := 0; y < height; y++ {
		var row strings.Builder
		row.WriteString(colorDim)
		row.WriteRune('|')
		row.WriteString(colorReset)
		for x := 0; x < width; x++ {
			row.WriteString(styleCell(grid[y][x]))
		}
		row.WriteString(colorDim)
		row.WriteRune('|')
		row.WriteString(colorReset)
		fmt.Println(row.String())
	}
	fmt.Printf("%s+%s+%s\n\n", colorDim, hLine, colorReset)
}

func printFooter(stats RenderStats) {
	dirs := []Direction{Up, Down, Left, Right}
	fmt.Printf("%sLane Queues%s  ", colorBold, colorReset)
	for _, d := range dirs {
		fmt.Printf("%s=%d  ", d, stats.LaneQueue[d])
	}
	fmt.Println()

	fmt.Printf("%sLane Active%s  ", colorBold, colorReset)
	for _, d := range dirs {
		fmt.Printf("%s=%d  ", d, stats.LaneActive[d])
	}
	fmt.Println()

	legend := []string{
		colorGreen + "G" + colorReset + "=vertical green",
		colorRed + "R" + colorReset + "=horizontal green",
		colorCyan + "^/v" + colorReset + "=vertical cars",
		colorYellow + "</>" + colorReset + "=horizontal cars",
		colorGray + "|/-" + colorReset + "=roads",
	}
	sort.Strings(legend)
	fmt.Printf("%sLegend%s %s\n", colorBold, colorReset, strings.Join(legend, "  "))
}

func styleCell(ch rune) string {
	switch ch {
	case 'G':
		return colorGreen + "G" + colorReset
	case 'R':
		return colorRed + "R" + colorReset
	case '^', 'v':
		return colorCyan + string(ch) + colorReset
	case '<', '>':
		return colorYellow + string(ch) + colorReset
	case '|', '-':
		return colorGray + string(ch) + colorReset
	case ' ':
		return " "
	default:
		return colorBlue + string(ch) + colorReset
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
