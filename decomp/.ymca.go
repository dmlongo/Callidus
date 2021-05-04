package decomp

import (
	"fmt"
	"time"

	"github.com/dmlongo/callidus/csp"
	"github.com/dmlongo/callidus/db"
)

type ymca struct {
	tree *Node
	sol  csp.Solution
	all  []csp.Solution

	dir *director
}

func (y *ymca) Solve() (csp.Solution, bool) {
	if y.sol == nil {
		if y.reduce(y.tree) {
			// TODO backtrack
			// measure time diff of back with und ohne fullyReduce
		} else {
			y.sol = csp.Solution{}
		}
	}
	if len(y.sol) == 0 {
		return y.sol, false
	} else {
		return y.sol, true
	}
}

func (y *ymca) AllSolutions() []csp.Solution {
	if y.all == nil {
		y.fullyReduce(y.tree)
		_, rel := y.joinUpwards(y.tree)

		fmt.Print("(Conversion from Relation to Solution... ")
		startConversion := time.Now()
		y.all = db.RelToSolutions(rel)
		fmt.Print("done in ", time.Since(startConversion), ") ")
	}
	return y.all
}

func (y *ymca) reduce(root *Node) bool {
	y.dir = &director{}
	y.dir.setup(y.tree)
	for _, ya := range y.dir.agents {
		go ya.run()
	}
	y.dir.all <- &startMsg{}
	msg := <-y.dir.results
	if msg, ok := msg.(*stopMsg); ok {
		return msg.outcome
	} else {
		panic("Unexpected message: " + String(msg))
	}
}

func (y *ymca) fullyReduce(root *Node) {}

func (y *ymca) joinUpwards(root *Node) ([]string, db.Relation) {
	return nil, nil
}

type director struct {
	//tree   *Node
	agents []*yAgent

	root   chan<- Message
	leaves chan<- Message
	all    chan<- Message

	results   <-chan Message
	numLeaves int

	//signal <-chan bool
	stop chan<- bool
}

func (d *director) setup(root *Node) {
	rootCh := make(chan Message)
	resultsCh := make(chan Message)
	stopCh := make(chan bool)
	//d.agents = make(map[int]chan<- Message)
	d.root = rootCh
	d.results = resultsCh
	d.numLeaves = 0
	d.stop = stopCh

	var leavesChs []chan<- Message
	var allChs []chan<- Message

	var toVisit []*Node
	var toSetup []*yAgent
	rootAgent := newYAgent(root, stopCh)
	rootAgent.inbox = rootCh
	//rootAgent.parentW = resultsCh
	rootAgent.director = resultsCh
	//rootAgent := &yAgent{inbox: rootCh, parentW: resultsCh, masterW: resultsCh}
	toVisit = append(toVisit, root)
	toSetup = append(toSetup, rootAgent)
	var parent *Node
	var parAgent *yAgent
	for len(toVisit) > 0 {
		parent, toVisit = toVisit[0], toVisit[1:]
		parAgent, toSetup = toSetup[0], toSetup[1:]
		d.agents = append(d.agents, parAgent)

		fromDirector := make(chan Message)
		if len(parent.Children) > 0 {
			var parWriteChs []chan<- Message
			var parReadChs []<-chan Message
			for _, child := range parent.Children {
				toChild := make(chan Message)
				fromChild := make(chan Message)
				// childAgent := &YAgent{id: child.ID, myRel: child.Table, myNode: child, state: waiting, inbox: toChild, parentW: fromChild, masterW: signal, stop: stop}
				childAgent := newYAgent(child, stopCh)
				childAgent.inbox = toChild
				childAgent.parentW = fromChild
				childAgent.director = resultsCh
				parWriteChs = append(parWriteChs, toChild)
				parReadChs = append(parReadChs, fromChild)

				toVisit = append(toVisit, child)
				toSetup = append(toSetup, childAgent)
			}
			parReadChs = append(parReadChs, fromDirector)
			parReadChs = append(parReadChs, parAgent.inbox)
			parAgent.inbox = merge(stopCh, parReadChs...)
			parAgent.childrenW = multicast(parWriteChs...)
		} else {
			parAgent.inbox = merge(stopCh, parAgent.inbox, fromDirector)
			leavesChs = append(leavesChs, fromDirector)
			//toDirector := make(chan Message)
			//parAgent.childrenW = toOutside
			//leavesR = append(leavesR, toOutside)
			parAgent.childrenW = resultsCh
			d.numLeaves++
		}
		allChs = append(allChs, fromDirector)
	}

	d.leaves = multicast(leavesChs...)
	d.all = multicast(allChs...)
}

type yAgent struct {
	myNode *Node
	state  int

	deferred chan Message
	inbox    <-chan Message
	stop     <-chan bool

	parentW   chan<- Message
	childrenW chan<- Message
	director  chan<- Message
}

func newYAgent(n *Node, stop <-chan bool) (ya *yAgent) {
	ya = &yAgent{}
	ya.myNode = n
	ya.state = waiting
	ya.stop = stop
	return
}

func (ya *yAgent) id() int {
	return ya.myNode.ID
}

func (ya *yAgent) rel() db.Relation {
	return ya.myNode.Table
}

func (ya *yAgent) setRel(rel db.Relation) {
	ya.myNode.Table = rel
}

func (ya *yAgent) isRoot() bool {
	return ya.myNode.Parent == nil
}

func (ya *yAgent) numChildren() int {
	return len(ya.myNode.Children)
}

func (ya *yAgent) run() {
	//defer ya.quit()
	cnt := 0
	lastState := ya.state
	var msg Message
	fmt.Println("Agent", ya.id(), "ready to go.")
	for {
		if ya.state != lastState {
			select {
			case msg = <-ya.deferred:
			}
		} else {
		select {
		case msg = <-ya.inbox:
		case <-ya.stop: // TODO deal with this early stop situation
			return
		}
		fmt.Println("Agent", ya.id(), "received", String(msg))

		lastState = ya.state
		switch ya.state {
		case waiting:
			if _, ok := msg.(*startMsg); ok {
				fmt.Println("Agent", ya.id(), "is computing.")
				// TODO compute my own rel
				if ya.rel().Empty() {
					select {
					case ya.director <- &stopMsg{false}:
					case <-ya.stop:
					}
					return
				}
				if ya.numChildren() == 0 {
					if ya.isRoot() { // single node
						ya.director <- &stopMsg{true}
						close(ya.director)
						ya.state = finished
					} else {
						ya.parentW <- &semijoinMsg{ya.rel()} // TODO could be filter
						ya.state = reduced
					}
				} else {
					ya.state = computed
				}
			} else {
				ya.postpone(msg)
			}
		case computed:
			_, sj := msg.(*semijoinMsg)
			_, sel := msg.(*selectMsg)
			if sj || sel {
				switch msg := msg.(type) {
				case *semijoinMsg:
					fmt.Println("Agent", ya.id(), "is semijoining from down.")
					othRel := msg.Content().(db.Relation)
					db.Semijoin(ya.rel(), othRel)
					cnt++
				case *selectMsg:
					cond := msg.Content().(db.Condition)
					db.Select(ya.rel(), cond)
					cnt++
				}
				if ya.rel().Empty() {
					select {
					case ya.director <- &stopMsg{false}: // todo might block other people sending semijoin
					case <-ya.stop:
					}
					return
				}
				if cnt == ya.numChildren() {
					cnt = 0
					if ya.isRoot() {
						ya.director <- &stopMsg{true}
						//close(ya.director)
						//ya.state = finished
					} else {
						ya.parentW <- &semijoinMsg{ya.rel()} // TODO could be filter
					}
					ya.state = reduced
				}
			} else {
				ya.postpone(msg)
			}
		case reduced:
			_, sj := msg.(*semijoinMsg)
			_, sel := msg.(*selectMsg)
			if sj || sel {
				switch msg := msg.(type) {
				case *semijoinMsg:
					fmt.Println("Agent", ya.id(), "is semijoining from up.")
					othRel := msg.Content().(db.Relation)
					db.Semijoin(ya.rel(), othRel)
				case *selectMsg:
					cond := msg.Content().(db.Condition)
					db.Select(ya.rel(), cond)
				}
				if ya.numChildren() > 0 {
					ya.childrenW <- &semijoinMsg{ya.rel()} // TODO could be filter
				} else {
					ya.director <- &stopMsg{true}
				}
				ya.state = fullyReduced
			} else {
				ya.postpone(msg)
			}
		case fullyReduced:
			if msg, ok := msg.(*joinMsg); ok {
				fmt.Println("Agent", ya.id(), "is joining.")
				othRel := msg.Content().(db.Relation)
				ya.setRel(db.Join(ya.rel(), othRel))
				cnt++
				if cnt == len(ya.myNode.Children) {
					cnt = 0
					ya.parentW <- &joinMsg{ya.rel()}
					ya.state = finished
				}
				if cnt == ya.numChildren() {
					cnt = 0
					if ya.isRoot() {
						ya.director <- &stopMsg{true}
						close(ya.director)
						ya.state = finished
					} else {
						ya.parentW <- &joinMsg{ya.rel()}
						ya.state = reduced
					}
				}
			} else {
				ya.postpone(msg)
			}
		case finished:
			// TODO
		case crashed:
		default:
			panic(fmt.Sprintf("Illegal state: %v", ya.state))
		}
	}
}

func (ya *yAgent) postpone(msg Message) {
	go func() { ya.deferred <- msg }()
}
