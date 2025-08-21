package tax

import "strconv"

type TaxRate struct {
	v uint8
}

var DefaultTaxRate = TaxRate{v: 10}

func (t TaxRate) String() string {
	return strconv.FormatUint(uint64(t.v), 10)
}

func (t TaxRate) Uint8() uint8 {
	return t.v
}
