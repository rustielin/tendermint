package consensus

func discardFromChan(ch chan interface{}, n int) {
	for i := 0; i < n; i++ {
		<-ch
	}
}
