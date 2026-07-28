// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package oracletest carries the two mode switches and the
// cross-check that runs a fixture through both engines.
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

// The pieces CrossCheck stands on. Extraction writes a fixture's
// input files into a directory; the Emacs side copies lit/ in beside
// them and runs the driver exactly as lit.mk does; the comparison
// walks both trees and reports the first byte that differs, or the
// first file one side has and the other lacks.
func extract(t *testing.T, fixture, dir string) (book string) {
	panic("HOLE(6): write the txtar's input files into dir, and name the book among them")
}

func runOracle(t *testing.T, dir, book string) {
	panic("HOLE(6): copy lit/ in, run batch Emacs with the driver, fail loudly when it is absent")
}

func compareTrees(t *testing.T, oracle, babble string) {
	panic("HOLE(6): every file on both sides, byte for byte, ignoring the inputs")
}

func remint(t *testing.T, fixture, oracle string) {
	panic("HOLE(6): rewrite the fixture's want/ files from the oracle tree")
}
