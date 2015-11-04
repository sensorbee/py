package mainthread

var (
	jobs = make(chan func())
)

// Exec asynchronously executes a function on the main thread. Callers generally
// use this function as follows:
//
//	ch := make(chan ResultType)
//	mainthread.Exec(func() {
//		// Do something related to Python
//		// DO NOT call a function calling mainthread.Exec in it from here
//		ch <- result
//	})
//	r := <- ch
//
// As noted in the comment above, a function passed to Exec must not call
// another function which also calls Exec. It will result in a deadlock because
// Exec isn't reentrant.
//
// This function panics after Terminate is called. However, callers can assume
// that this will never panic because Terminate is only provided for debugging
// purpose and will not be used in a production envinronment.
func Exec(f func()) {
	jobs <- f
}

// ExecSync is the synchronous version of Exec. It waits until f finishes.
// This function is useful when f doesn't have to return a value.
func ExecSync(f func()) {
	ch := make(chan struct{})
	Exec(func() {
		defer func() {
			ch <- struct{}{}
		}()
		f()
	})
	<-ch
}

func process() {
	for f := range jobs {
		func() {
			defer recover()
			f()
		}()
	}
}

// Terminate terminates the main thread. After calling this function, Exec,
// ExecSync, and other Python modules for SensorBee will no longer work.
// This function is basically provided for debugging purpose. More specifically,
// it's used to call Py_Finalize.
func Terminate() {
	close(jobs)
}
