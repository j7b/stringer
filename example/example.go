package example

type Int int

const (
	_       = iota
	One Int = iota
	Two
)

type Comparison int

const (
	Equal Comparison = iota
	Greater
	Less Comparison = -1
)

const Four = Int(4)

const (
	Weird       string = "weird"
	Pie                = 22 / 7
	NotThreeTwo Int    = iota
)
