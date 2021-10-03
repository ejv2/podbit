// Lists your configured/detected podcasts and available episodes
package ui

type List struct {
	test string
}

func (l List) Name() string {
	return "Podcasts"
}

func (l List) Render(x, y int) {
	root.MovePrint(y, x, "The list will be here")
}
