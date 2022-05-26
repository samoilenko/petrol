package infrastructure

func runPipe(pipe Job, in chan interface{}, c chan int) chan interface{} {
	out := make(chan interface{})
	go func() {
		pipe(in, out)
		close(out)
		c <- 1
	}()

	return out
}

// ExecutePipeline this function works like Linux pipeline
func ExecutePipeline(pipes ...Job) {
	counter := make(chan int, len(pipes))
	var in chan interface{}
	for _, pipe := range pipes {
		in = runPipe(pipe, in, counter)
	}

	for range pipes {
		<-counter
	}

	return
}

type Job func(in, out chan interface{})
