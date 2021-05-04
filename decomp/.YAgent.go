package decomp

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dmlongo/callidus/csp"
	"github.com/dmlongo/callidus/db"
)

// Message to be processed by YAgents
type Message interface {
	Content() interface{}
}

func String(msg Message) string {
	return fmt.Sprintf("msg= %T : %v", msg, msg.Content())
}

type startMsg struct{}

func (msg *startMsg) Content() interface{} { return nil }

type stopMsg struct {
	outcome bool
}

func (msg *stopMsg) Content() interface{} { return msg.outcome }

type semijoinMsg struct {
	rel db.Relation
}

func (msg *semijoinMsg) Content() interface{} {
	return msg.rel
}

type selectMsg struct {
	c *db.Condition
}

func (msg *selectMsg) Content() interface{} {
	return msg.c
}

type joinMsg struct {
	rel db.Relation
}

func (msg *joinMsg) Content() interface{} {
	return msg.rel
}

const (
	waiting = iota
	computed
	reduced
	fullyReduced
	finished
	crashed
)

type YAgent struct {
	id     int
	myRel  db.Relation
	myNode *Node
	state  int

	inbox <-chan Message
	stop  <-chan bool

	masterW   chan<- bool
	parentW   chan<- Message
	childrenW chan<- Message
}

func (ya *YAgent) Run() {
	defer ya.quit()
	cnt := 0
	var msg Message
	fmt.Println("Agent", ya.id, "ready to go.")
	for {
		select {
		case msg = <-ya.inbox:
		case <-ya.stop: // TODO deal with this early stop situation
			return
		}

		fmt.Println("Agent", ya.id, "received", String(msg))

		switch ya.state {
		case waiting:
			fmt.Println("Agent", ya.id, "is computing.")
			// TODO compute my own rel
			if ya.myRel.Empty() {
				select {
				case ya.masterW <- false:
					close(ya.masterW)
				default:
				}
				return
			}
			if len(ya.myNode.Children) == 0 {
				ya.parentW <- &semijoinMsg{rel: ya.myRel} // TODO could be filter
				ya.state = reduced
			} else {
				ya.state = computed
			}
		case computed:
			switch msg := msg.(type) {
			case *semijoinMsg:
				fmt.Println("Agent", ya.id, "is semijoining from down.")
				othRel := msg.Content().(db.Relation)
				ya.myRel, _ = db.Semijoin(ya.myRel, othRel)
				cnt++
			case *selectMsg:
				cond := msg.Content().(db.Condition)
				ya.myRel, _ = db.Select(ya.myRel, cond)
				cnt++
			}
			if ya.myRel.Empty() {
				select {
				case ya.masterW <- false:
					close(ya.masterW)
				default:
				}
				return
			}
			if cnt == len(ya.myNode.Children) {
				cnt = 0
				ya.parentW <- &semijoinMsg{rel: ya.myRel} // TODO could be filter
				ya.state = reduced
			}
		case reduced:
			switch msg := msg.(type) {
			case *semijoinMsg:
				fmt.Println("Agent", ya.id, "is semijoining from up.")
				othRel := msg.Content().(db.Relation)
				ya.myRel, _ = db.Semijoin(ya.myRel, othRel)
			case *selectMsg:
				cond := msg.Content().(db.Condition)
				ya.myRel, _ = db.Select(ya.myRel, cond)
			}
			ya.childrenW <- &semijoinMsg{rel: ya.myRel} // TODO could be filter
			ya.state = fullyReduced
		case fullyReduced:
			switch msg := msg.(type) {
			case *joinMsg:
				fmt.Println("Agent", ya.id, "is joining.")
				othRel := msg.Content().(db.Relation)
				ya.myRel = db.Join(ya.myRel, othRel)
				cnt++
			}
			if cnt == len(ya.myNode.Children) {
				cnt = 0
				ya.parentW <- &joinMsg{rel: ya.myRel}
				ya.state = finished
			}
		case finished:
			// TODO
		case crashed:
		default:
			panic(fmt.Sprintf("Illegal state: %v", ya.state))
		}
	}
}

func (ya *YAgent) quit() {
	close(ya.parentW)
	close(ya.childrenW)
	ya.state = finished
}

type YPipeline struct {
	root      chan<- Message
	leaves    chan<- Message
	broadcast chan<- Message

	result <-chan Message

	signal <-chan bool
	stop   chan<- bool

	state     int
	numLeaves int

	agents []*YAgent
}

func SetupPipeline(root *Node) YPipeline {
	toRoot, fromRoot, toLeaves, broadcast, signal, stop, agents, numLeaves := setup(root)
	return YPipeline{root: toRoot, leaves: toLeaves, broadcast: broadcast, signal: signal, stop: stop, result: fromRoot, state: waiting, agents: agents, numLeaves: numLeaves}
}

func setup(root *Node) (toRoot chan Message, fromRoot <-chan Message, toLeaves chan<- Message, broadcast chan<- Message, signal chan bool, stop chan bool, agents []*YAgent, numLeaves int) {
	numLeaves = 0
	toRoot = make(chan Message)
	fromRootOutside := make(chan Message)
	signal = make(chan bool)
	stop = make(chan bool)
	var leavesW []chan<- Message
	var leavesR []<-chan Message
	var all []chan<- Message

	var toVisit []*Node
	var toSetup []*YAgent
	rootAgent := &YAgent{id: root.ID, myRel: root.Table, myNode: root, state: waiting, inbox: toRoot, parentW: fromRootOutside, masterW: signal, stop: stop}
	toVisit = append(toVisit, root)
	toSetup = append(toSetup, rootAgent)
	var parent *Node
	var parAgent *YAgent
	for len(toVisit) > 0 {
		parent, toVisit = toVisit[0], toVisit[1:]
		parAgent, toSetup = toSetup[0], toSetup[1:]
		agents = append(agents, parAgent)

		fromMaster := make(chan Message)
		if len(parent.Children) > 0 {
			var toChildren []chan<- Message
			var fromChildren []<-chan Message
			for _, child := range parent.Children {
				toChild := make(chan Message)
				fromChild := make(chan Message)
				childAgent := &YAgent{id: child.ID, myRel: child.Table, myNode: child, state: waiting, inbox: toChild, parentW: fromChild, masterW: signal, stop: stop}
				toChildren = append(toChildren, toChild)
				fromChildren = append(fromChildren, fromChild)

				toVisit = append(toVisit, child)
				toSetup = append(toSetup, childAgent)
			}
			fromChildren = append(fromChildren, fromMaster)
			fromChildren = append(fromChildren, parAgent.inbox)
			parAgent.inbox = merge(stop, fromChildren...)
			parAgent.childrenW = multicast(toChildren...)
		} else {
			fromOutside := make(chan Message)
			parAgent.inbox = merge(stop, parAgent.inbox, fromOutside, fromMaster)
			leavesW = append(leavesW, fromOutside)
			toOutside := make(chan Message)
			parAgent.childrenW = toOutside
			leavesR = append(leavesR, toOutside)
			numLeaves++
		}
		all = append(all, fromMaster)
	}

	toLeaves = multicast(leavesW...)
	broadcast = multicast(all...)
	leavesR = append(leavesR, fromRootOutside)
	fromRoot = merge(stop, leavesR...)
	return
}

func (yp *YPipeline) Sat() (bool, error) {
	if yp.state != waiting {
		return false, errors.New("Illegal state: " + strconv.Itoa(yp.state))
	}

	for _, ya := range yp.agents {
		go ya.Run()
	}
	yp.broadcast <- &startMsg{}

	select {
	case msg := <-yp.result:
		if msg, ok := msg.(*semijoinMsg); ok && msg.Content() != nil {
			yp.state = computed
			return true, nil
		} else {
			yp.state = crashed
			return false, errors.New("Unexpected message: " + String(msg))
		}
	case <-yp.signal:
		yp.stop <- true
		// or close(yp.stop)?
		return false, nil
	}
}

// TODO func (yp *YPipeline) One() (ctr.Solution, error) - it gives you one solution

func (yp *YPipeline) Reduce() error {
	if yp.state != computed {
		return errors.New("Illegal state: " + strconv.Itoa(yp.state))
	}

	cnt := 0
	yp.root <- &startMsg{}

	msg := <-yp.result
	if msg, ok := msg.(*semijoinMsg); ok {
		cnt++
		if cnt == yp.numLeaves {
			return nil
		}
	} else {
		yp.state = crashed
		return errors.New("Unexpected message: " + String(msg))
	}
	return nil
}

func (yp *YPipeline) All() ([]csp.Solution, error) { // TODO check with the new architecture
	if yp.state != computed {
		return nil, errors.New("Illegal state: " + strconv.Itoa(yp.state))
	}

	yp.root <- &startMsg{}

	msg := <-yp.result
	if msg, ok := msg.(*joinMsg); ok {
		rel := msg.Content().(db.Relation)
		yp.state = finished
		return db.RelToSolutions(rel), nil
	} else {
		yp.state = crashed
		return nil, errors.New("Unexpected message: " + String(msg))
	}
}

var ypipe YPipeline

func YMCA(root *Node) (bool, error) {
	ypipe = SetupPipeline(root)
	return ypipe.Sat()
}

func YMCAFullReduce() error {
	return ypipe.Reduce()
}

func YMCASols() ([]csp.Solution, error) {
	return ypipe.All()
}
