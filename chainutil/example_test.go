package chainutil_test

import (
	"fmt"
	"math"

	"github.com/flokiorg/go-flokicoin/chainutil"
)

func ExampleAmount() {

	a := chainutil.Amount(0)
	fmt.Println("Zero Loki:", a)

	a = chainutil.Amount(1e8)
	fmt.Println("100,000,000 Lokis:", a)

	a = chainutil.Amount(1e5)
	fmt.Println("100,000 Lokis:", a)
	// Output:
	// Zero Loki: 0 FLC
	// 100,000,000 Lokis: 1 FLC
	// 100,000 Lokis: 0.00100000 FLC
}

func ExampleNewAmount() {
	amountOne, err := chainutil.NewAmount(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountOne) //Output 1

	amountFraction, err := chainutil.NewAmount(0.01234567)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountFraction) //Output 2

	amountZero, err := chainutil.NewAmount(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountZero) //Output 3

	amountNaN, err := chainutil.NewAmount(math.NaN())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountNaN) //Output 4

	// Output: 1 FLC
	// 0.01234567 FLC
	// 0 FLC
	// invalid flokicoin amount
}

func ExampleAmount_unitConversions() {
	amount := chainutil.Amount(44433322211100)

	fmt.Println("Loki to kFLC:", amount.Format(chainutil.AmountKiloFLC))
	fmt.Println("Loki to FLC:", amount)
	fmt.Println("Loki to MilliFLC:", amount.Format(chainutil.AmountMilliFLC))
	fmt.Println("Loki to MicroFLC:", amount.Format(chainutil.AmountMicroFLC))
	fmt.Println("Loki to Loki:", amount.Format(chainutil.AmountLoki))

	// Output:
	// Loki to kFLC: 444.333222111 kFLC
	// Loki to FLC: 444333.22211100 FLC
	// Loki to MilliFLC: 444333222.111 mFLC
	// Loki to MicroFLC: 444333222111 μFLC
	// Loki to Loki: 44433322211100 Loki
}
