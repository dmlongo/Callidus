package decomp

import "testing"

func TestYMCA1(t *testing.T) {
	input, partial, output, _ := test1Data()
	sat, err := YMCA(input)
	if err != nil {
		panic(err)
	}
	if !sat {
		t.Error("y(input) is unsat, expected sat")
	}
	if !equals(input, partial) {
		t.Error("y(input) != partial")
	}

	err = YMCAFullReduce()
	if err != nil {
		panic(err)
	}
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	//if !equals(FullyReduceRelationsPar(partial), output) {
	//	t.Error("y(partial) != output")
	//}
	//if !solEquals(ComputeAllSolutions(output), sols) {
	//	t.Error("y(output) != solutions")
	//}
}
