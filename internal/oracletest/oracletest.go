// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package oracletest is the seam stage 6 fills: the two mode
// switches and the cross-check that runs a fixture through both
// engines.
package oracletest

import "testing"

func Enabled() bool {
	panic("HOLE(6): ORACLE=1 in the environment")
}

func Reminting() bool {
	panic("HOLE(6): UPDATE=1 in the environment")
}

func CrossCheck(t *testing.T, fixture string) {
	panic("HOLE(6): extract the fixture twice, run Emacs and babble, byte-compare the trees")
}
