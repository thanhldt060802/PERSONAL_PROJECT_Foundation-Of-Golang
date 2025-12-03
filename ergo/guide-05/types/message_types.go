package types

import "ergo.services/ergo/net/edf"

type SimpleMessage struct {
	Data string
}

// Register network messages for interacting with another node
func init() {
	if err := edf.RegisterTypeOf(SimpleMessage{}); err != nil {
		panic(err)
	}
}
