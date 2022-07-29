package colors

const (
	// Smallest iota color value (practically always 1).
	colorMin = ColorRed
	// Largest iota color value.
	colorMax = BackgroundCyan
	// Distance between a foreground color and a background color.
	colorBoundary = colorMax / 2
)

// ToForeground returns the passed color converted such that a background color
// is its foreground equivalent.
func ToForeground(color int) int {
	if color <= colorBoundary {
		return color
	}

	return color - colorBoundary
}

// ToForeground returns the passed color converted such that a foreground color
// is its background equivalent.
func ToBackground(color int) int {
	if color > colorBoundary {
		return color
	}

	return color + colorBoundary
}
