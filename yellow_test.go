package yellow

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"
)

type chanTimedOutHandler struct{ ch, ch2 chan bool }

func (n *chanTimedOutHandler) TimedOut(t time.Time)  { close(n.ch) }
func (n *chanTimedOutHandler) Completed(t time.Time) { close(n.ch2) }
func (n *chanTimedOutHandler) Wait()                 { <-n.ch }
func (n *chanTimedOutHandler) WaitComplete()         { <-n.ch2 }

type failHandler struct{ t *testing.T }

func (n *failHandler) TimedOut(t time.Time)  { n.t.Fatalf("Unexpected timeout") }
func (n *failHandler) Completed(t time.Time) { n.t.Fatalf("Unexpected completed") }

func mkChanTimedOutHandler() *chanTimedOutHandler {
	return &chanTimedOutHandler{make(chan bool), make(chan bool)}
}

func TestNoDuration(t *testing.T) {
	s := Deadline(0, &failHandler{t})
	if s != nil {
		t.Fatalf("Expected nil stopwith 0 duration: %v", s)
	}
	s.Done() // noop
}

func TestNoDurationLog(t *testing.T) {
	s := DeadlineLog(0, "nope")
	if s != nil {
		t.Fatalf("Expected nil stopwith 0 duration: %v", s)
	}
	s.Done() // noop
}

func assertTimedOut(t *testing.T, n interface{}) {
	var ch chan bool
	switch h := n.(type) {
	case *chanTimedOutHandler:
		ch = h.ch
	case *chanHandler:
		ch = h.ch
	default:
		t.Fatalf("Unhandled type: %T", n)
	}
	select {
	case <-ch:
	default:
		t.Fatalf("Expected timeout, but didn't get one")
	}
}

func TestTimeoutWarning(t *testing.T) {
	ch := mkChanTimedOutHandler()
	defer assertTimedOut(t, ch)
	defer Deadline(1, ch).Done()
	<-ch.ch
}

func TestNoTimeoutWarning(t *testing.T) {
	defer Deadline(time.Minute, &failHandler{t}).Done()
}

func TestLoggerWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)
	lh := logHandler{"got %q", []interface{}{"x"}}
	lh.TimedOut(time.Now())
	lh.Completed(time.Now())
	// Should probably actually inspect this stuff.
	t.Logf("%s", buf.Bytes())
}

type chanHandler struct{ ch chan bool }

func (n *chanHandler) Completed(t time.Time) { close(n.ch) }
func (n *chanHandler) Wait()                 { <-n.ch }

func mkChanHandler() *chanHandler {
	return &chanHandler{make(chan bool)}
}

func TestTimeoutNoWarning(t *testing.T) {
	ch := mkChanHandler()
	defer assertTimedOut(t, ch)
	defer Deadline(1, ch).Done()
	time.Sleep(time.Millisecond)
}

func TestNoTimeoutNoWarning(t *testing.T) {
	Deadline(time.Minute, &failHandler{t}).Done()
}

func TestHandlerFunc(t *testing.T) {
	worked := false
	func() {
		defer Deadline(1, HandleFunc(func(time.Time) { worked = true })).Done()
		time.Sleep(time.Millisecond)
	}()
	if worked == false {
		t.Errorf("Failed to signal from HandleFunc")
	}
}

func BenchmarkNoDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Deadline(0, nil).Done()
	}
}

func BenchmarkMSDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Deadline(time.Millisecond, nil).Done()
	}
}

func BenchmarkNSDurationRef(b *testing.B) {
	refTime := time.Now()
	for i := 0; i < b.N; i++ {
		n := mkChanTimedOutHandler()
		n.TimedOut(refTime)
		n.Wait()
		n.Completed(refTime)
	}
}

func BenchmarkNSDurationWarningNotFired(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := mkChanTimedOutHandler()
		y := Deadline(1, n)
		y.Done()
	}
}

func BenchmarkNSDurationWarningFired(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := mkChanTimedOutHandler()
		y := Deadline(1, n)
		n.Wait()
		y.Done()
	}
}

func BenchmarkNSNoWarningDurationFired(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := mkChanHandler()
		y := Deadline(1, n)
		y.Done()
	}
}

func ExampleDeadline() {
	var myHandler Handler // presumably from somewhere
	defer Deadline(time.Second, myHandler).Done()
	// do something that should take less than a second
}

func ExampleDeadlineLog() {
	defer DeadlineLog(time.Second, "Doing thing %d", 1).Done()
	// do something that should take less than a second, log
	// otherwise
}
