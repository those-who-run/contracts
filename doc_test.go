package contracts_test

import (
	"fmt"

	"github.com/those-who-run/contracts"
)

func ExampleName() {
	fmt.Println(contracts.Name())
	// Output: contracts
}
