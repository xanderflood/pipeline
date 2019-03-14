package pipeline

//StartCounter starts a goroutine that listens to `signal` until
//it has either been closed, or until it has received `target`
//number of messages. At that point, `finished` will receive a
//single message and close.
func StartCounter(target int) (signal chan<- struct{}, finished <-chan struct{}) {
	signalCh := make(chan struct{}, target)
	finishedCh := make(chan struct{}, 1)

	go func() {
		defer close(finishedCh)
		defer func() { finishedCh <- struct{}{} }()

		count := 0
		for range signalCh {
			count++
			if count >= target {
				break
			}
		}
	}()

	return signalCh, finishedCh
}
