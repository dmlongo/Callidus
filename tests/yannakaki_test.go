package tests

import (
	. "../../Callidus/computation"
	. "../../Callidus/hyperTree"
	"reflect"
	"testing"
)

func Test1(t *testing.T) {
	input, output := test1Data()
	input = Yannakaki(input, false)
	if !reflect.DeepEqual(input.PossibleValues, output.PossibleValues) {
		t.Error()
	}
}

func Test2(t *testing.T) {
	input, output := test2Data()
	input = Yannakaki(input, false)
	if !reflect.DeepEqual(input.PossibleValues, output.PossibleValues) {
		t.Error()
	}
}

func test1Data() (*Node, *Node) {
	//creating input
	dInput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: map[string][]string{"Y": {"3", "3", "5", "6"}, "P": {"8", "7", "7", "7"}}}
	rInput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: map[string][]string{"Y": {"3", "9", "8", "3", "3", "8", "9"}, "Z": {"8", "3", "3", "8", "8", "9", "4"}, "U": {"9", "8", "8", "4", "3", "4", "7"}}}
	sInput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: map[string][]string{"Z": {"3", "9", "8", "3", "3", "8", "9"}, "U": {"8", "3", "3", "8", "8", "9", "4"}, "W": {"9", "8", "8", "4", "3", "4", "7"}}}
	tInput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: map[string][]string{"V": {"9", "9", "9"}, "Z": {"8", "3", "5"}}}
	dInput.AddSon(rInput)
	rInput.AddSon(sInput)
	rInput.AddSon(tInput)

	//creating output
	dOutput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: map[string][]string{"Y": {"3", "3"}, "P": {"8", "7"}}}
	rOutput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: map[string][]string{"Y": {"3", "3"}, "Z": {"8", "8"}, "U": {"9", "3"}}}
	sOutput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: map[string][]string{"Z": {"8", "8"}, "U": {"3", "9"}, "W": {"8", "4"}}}
	tOutput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: map[string][]string{"V": {"9"}, "Z": {"8"}}}
	dOutput.AddSon(rOutput)
	rOutput.AddSon(sOutput)
	rOutput.AddSon(tOutput)

	return dInput, dOutput
}

func test2Data() (*Node, *Node) {
	//creating input
	dInput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: map[string][]string{"Y": {"3", "3", "5", "6"}, "P": {"8", "7", "7", "7"}}}
	rInput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: map[string][]string{"Y": {"3", "9", "8", "3", "3", "8", "9"}, "Z": {"8", "3", "3", "8", "8", "9", "4"}, "U": {"9", "8", "8", "4", "3", "4", "7"}}}
	sInput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: map[string][]string{"Z": {"3", "9", "8", "3", "3", "8", "9"}, "U": {"8", "3", "3", "8", "8", "9", "4"}, "W": {"9", "8", "8", "4", "3", "4", "7"}}}
	tInput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: map[string][]string{"V": {"9", "9", "9"}, "Z": {"8", "3", "5"}}}
	aInput := &Node{Id: 5, Variables: []string{"P", "C"}, PossibleValues: map[string][]string{"P": {"8", "8", "4", "3"}, "C": {"4", "7", "9", "5"}}}
	bInput := &Node{Id: 6, Variables: []string{"C", "A"}, PossibleValues: map[string][]string{"C": {"4", "3", "5"}, "A": {"1", "2", "4"}}}
	dInput.AddSon(rInput)
	dInput.AddSon(aInput)
	rInput.AddSon(sInput)
	rInput.AddSon(tInput)
	aInput.AddSon(bInput)

	//creating output
	dOutput := &Node{Id: 1, Variables: []string{"Y", "P"}, PossibleValues: map[string][]string{"Y": {"3"}, "P": {"8"}}}
	rOutput := &Node{Id: 2, Variables: []string{"Y", "Z", "U"}, PossibleValues: map[string][]string{"Y": {"3", "3"}, "Z": {"8", "8"}, "U": {"9", "3"}}}
	sOutput := &Node{Id: 3, Variables: []string{"Z", "U", "W"}, PossibleValues: map[string][]string{"Z": {"8", "8"}, "U": {"3", "9"}, "W": {"8", "4"}}}
	tOutput := &Node{Id: 4, Variables: []string{"V", "Z"}, PossibleValues: map[string][]string{"V": {"9"}, "Z": {"8"}}}
	aOutput := &Node{Id: 5, Variables: []string{"P", "C"}, PossibleValues: map[string][]string{"P": {"8"}, "C": {"4"}}}
	bOutput := &Node{Id: 6, Variables: []string{"C", "A"}, PossibleValues: map[string][]string{"C": {"4"}, "A": {"1"}}}
	dOutput.AddSon(rOutput)
	dOutput.AddSon(aOutput)
	rOutput.AddSon(sOutput)
	rOutput.AddSon(tOutput)
	aOutput.AddSon(bOutput)

	return dInput, dOutput
}
