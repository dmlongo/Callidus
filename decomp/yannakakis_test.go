package decomp

import (
	"testing"
)

func TestYannakSeq1(t *testing.T) {
	input, partial, output, sols := test1Data()
	y, _ := NewYannakakis(input, "seq")
	if sat := y.reduce(input); !sat || !equals(input, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	y.fullyReduce(input)
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(y.AllSolutions(), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestYannakSeq2(t *testing.T) {
	input, partial, output, sols := test2Data()
	y, _ := NewYannakakis(input, "seq")
	if sat := y.reduce(input); !sat || !equals(input, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	y.fullyReduce(input)
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(y.AllSolutions(), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestYannakSeq3(t *testing.T) {
	input := test3Data()
	if sat := (&seqY{}).reduce(input); sat {
		t.Error("y(input) is sat!")
	}
}

func TestYannakPar1(t *testing.T) {
	input, partial, output, sols := test1Data()
	y, _ := NewYannakakis(input, "par")
	if sat := y.reduce(input); !sat || !equals(input, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	y.fullyReduce(input)
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(y.AllSolutions(), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestYannakPar2(t *testing.T) {
	input, partial, output, sols := test2Data()
	y, _ := NewYannakakis(input, "par")
	if sat := y.reduce(input); !sat || !equals(input, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	y.fullyReduce(input)
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(y.AllSolutions(), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestYannakPar3(t *testing.T) {
	input := test3Data()
	if sat := (&parY{}).reduce(input); sat {
		t.Error("y(input) is sat!")
	}
}

func TestYannakYMCA1(t *testing.T) {
	input, partial, output, sols := test1Data()
	y, _ := NewYannakakis(input, "ymca")
	if sat := y.reduce(input); !sat || !equals(input, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	y.fullyReduce(input)
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(y.AllSolutions(), sols) {
		t.Error("y(output) != solutions")
	}
}

func TestYannakYMCA2(t *testing.T) {
	input, partial, output, sols := test2Data()
	y, _ := NewYannakakis(input, "ymca")
	if sat := y.reduce(input); !sat || !equals(input, partial) {
		if !sat {
			t.Error("y(input) is unsat!")
		}
		t.Error("y(input) != partial")
	}
	y.fullyReduce(input)
	if !equals(input, output) {
		t.Error("y(partial) != output")
	}
	if !solEquals(y.AllSolutions(), sols) {
		t.Error("y(output) != solutions")
	}
}

/*func TestYannakYMCA3(t *testing.T) {
	input := test3Data()
	if sat := (&ymca{}).reduce(input); sat {
		t.Error("y(input) is sat!")
	}
}*/
