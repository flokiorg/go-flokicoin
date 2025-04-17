// Copyright (c) 2013, 2014 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainutil

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// AmountUnit describes a method of converting an Amount to something
// other than the base unit of a flokicoin.  The value of the AmountUnit
// is the exponent component of the decadic multiple to convert from
// an amount in flokicoin to an amount counted in units.
type AmountUnit int

// These constants define various units used when describing a flokicoin
// monetary amount.
const (
	AmountMegaFLC  AmountUnit = 6
	AmountKiloFLC  AmountUnit = 3
	AmountFLC      AmountUnit = 0
	AmountMilliFLC AmountUnit = -3
	AmountMicroFLC AmountUnit = -6
	AmountLoki     AmountUnit = -8
)

// String returns the unit as a string.  For recognized units, the SI
// prefix is used, or "Loki" for the base unit.  For all unrecognized
// units, "1eN FLC" is returned, where N is the AmountUnit.
func (u AmountUnit) String() string {
	switch u {
	case AmountMegaFLC:
		return "MFLC"
	case AmountKiloFLC:
		return "kFLC"
	case AmountFLC:
		return "FLC"
	case AmountMilliFLC:
		return "mFLC"
	case AmountMicroFLC:
		return "μFLC"
	case AmountLoki:
		return "Loki"
	default:
		return "1e" + strconv.FormatInt(int64(u), 10) + " FLC"
	}
}

// Amount represents the base flokicoin monetary unit (colloquially referred
// to as a `Loki').  A single Amount is equal to 1e-8 of a flokicoin.
type Amount int64

// round converts a floating point number, which may or may not be representable
// as an integer, to the Amount integer type by rounding to the nearest integer.
// This is performed by adding or subtracting 0.5 depending on the sign, and
// relying on integer truncation to round the value to the nearest Amount.
func round(f float64) Amount {
	if f < 0 {
		return Amount(f - 0.5)
	}
	return Amount(f + 0.5)
}

// NewAmount creates an Amount from a floating point value representing
// some value in flokicoin.  NewAmount errors if f is NaN or +-Infinity, but
// does not check that the amount is within the total amount of flokicoin
// producible as f may not refer to an amount at a single moment in time.
//
// NewAmount is for specifically for converting FLC to Loki.
// For creating a new Amount with an int64 value which denotes a quantity of Loki,
// do a simple type conversion from type int64 to Amount.
// See GoDoc for example: http://godoc.org/github.com/flokiorg/go-flokicoin/chainutil#example-Amount
func NewAmount(f float64) (Amount, error) {
	// The amount is only considered invalid if it cannot be represented
	// as an integer type.  This may happen if f is NaN or +-Infinity.
	switch {
	case math.IsNaN(f):
		fallthrough
	case math.IsInf(f, 1):
		fallthrough
	case math.IsInf(f, -1):
		return 0, errors.New("invalid flokicoin amount")
	}

	return round(f * LokiPerFlokicoin), nil
}

// ToUnit converts a monetary amount counted in flokicoin base units to a
// floating point value representing an amount of flokicoin.
func (a Amount) ToUnit(u AmountUnit) float64 {
	return float64(a) / math.Pow10(int(u+8))
}

// ToFLC is the equivalent of calling ToUnit with AmountFLC.
func (a Amount) ToFLC() float64 {
	return a.ToUnit(AmountFLC)
}

// Format formats a monetary amount counted in flokicoin base units as a
// string for a given unit.  The conversion will succeed for any unit,
// however, known units will be formatted with an appended label describing
// the units with SI notation, or "Loki" for the base unit.
func (a Amount) Format(u AmountUnit) string {
	units := " " + u.String()
	formatted := strconv.FormatFloat(a.ToUnit(u), 'f', -int(u+8), 64)

	// When formatting full FLC, add trailing zeroes for numbers
	// with decimal point to ease reading of sat amount.
	if u == AmountFLC {
		if strings.Contains(formatted, ".") {
			return fmt.Sprintf("%.8f%s", a.ToUnit(u), units)
		}
	}
	return formatted + units
}

// String is the equivalent of calling Format with AmountFLC.
func (a Amount) String() string {
	return a.Format(AmountFLC)
}

// MulF64 multiplies an Amount by a floating point value.  While this is not
// an operation that must typically be done by a full node or wallet, it is
// useful for services that build on top of flokicoin (for example, calculating
// a fee by multiplying by a percentage).
func (a Amount) MulF64(f float64) Amount {
	return round(float64(a) * f)
}
