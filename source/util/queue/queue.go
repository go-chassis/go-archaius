package queue

import (
	"bytes"
	"errors"
	"sync"
)

// Concurrent is a framework that allows for concurrent N independent pieces of work
func Concurrent(workers, pieces int, doWorkPiece func(piece int, errCh chan error)) error {
	if pieces < workers {
		workers = pieces
	}
	var errList []error
	toProcess := make(chan int, pieces)
	errCh := make(chan error, pieces)
	for i := 0; i < pieces; i++ {
		toProcess <- i
	}
	close(toProcess)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(errCh chan error) {
			defer wg.Done()
			for piece := range toProcess {
				doWorkPiece(piece, errCh)
			}
		}(errCh)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		errList = append(errList, err)
	}
	if len(errList) > 0 {
		var errs bytes.Buffer
		for _, i := range errList {
			errs.WriteString("\t" + i.Error() + "\n")
		}
		return errors.New(errs.String())
	}
	return nil
}
