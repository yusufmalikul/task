package task

import (
	. "github.com/charmbracelet/lipgloss"
)

// Colors.
var (
	normal          = AdaptiveColor{Light: "#1A1A1A", Dark: "#dddddd"}
	normalDim       = AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}
	gray            = AdaptiveColor{Light: "#909090", Dark: "#626262"}
	midGray         = AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}
	darkGray        = AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
	brightGray      = AdaptiveColor{Light: "#847A85", Dark: "#979797"}
	dimBrightGray   = AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}
	indigo          = AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	dimIndigo       = AdaptiveColor{Light: "#9498FF", Dark: "#494690"}
	subtleIndigo    = AdaptiveColor{Light: "#7D79F6", Dark: "#514DC1"}
	dimSubtleIndigo = AdaptiveColor{Light: "#BBBDFF", Dark: "#383584"}
	cream           = AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}
	yellowGreen     = AdaptiveColor{Light: "#04B575", Dark: "#ECFD65"}
	dullYellowGreen = AdaptiveColor{Light: "#6BCB94", Dark: "#9BA92F"}
	fuschia         = AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	dimFuchsia      = AdaptiveColor{Light: "#F1A8FF", Dark: "#99519E"}
	dullFuchsia     = AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"}
	dimDullFuchsia  = AdaptiveColor{Light: "#F6C9FF", Dark: "#6B3A6F"}
	green           = Color("#04B575")
	red             = AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}
	faintRed        = AdaptiveColor{Light: "#FF6F91", Dark: "#C74665"}

	semiDimGreen = AdaptiveColor{Light: "#35D79C", Dark: "#036B46"}
	dimGreen     = AdaptiveColor{Light: "#72D2B0", Dark: "#0B5137"}
)

var (
	errorTitleStyle = NewStyle().
			Foreground(cream).
			Background(red).
			Padding(0, 1)

	subtleStyle = NewStyle().
			Foreground(AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
)
