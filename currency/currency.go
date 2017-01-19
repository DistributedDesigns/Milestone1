package currency


// Currency : Tracks dollars and cents as ints

type Currency struct {
	Dollars int
	Cents int
}

func (curr Currency) add(c1 Currency) {
	curr.Dollars += c1.Dollars + (curr.Cents + c1.Cents) /  100
	curr.Cents += (curr.Cents + c1.Cents) % 100
}