package decomp

import "sync"

func merge(done <-chan bool, cs ...<-chan Message) <-chan Message {
	var wg sync.WaitGroup
	out := make(chan Message)

	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan Message) {
			defer wg.Done()
			for msg := range c {
				select {
				case out <- msg:
				case <-done:
					return
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func multicast(cs ...chan<- Message) chan<- Message {
	in := make(chan Message)

	go func() {
		for msg := range in {
			for _, c := range cs {
				go func(c chan<- Message) { c <- msg }(c)
			}
		}
		for _, c := range cs {
			close(c)
		}
	}()

	return in
}
