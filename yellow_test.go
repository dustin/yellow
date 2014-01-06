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

func TestNoDurationLogWarn(t *testing.T) {
	s := DeadlineLogWarn(0, "nope")
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

var (
	_ = Handler(HandleFunc(func(time.Time) {}))
	_ = Handler(logHandler{})
	_ = Handler(logWarningHandler{})
	_ = TimedOutHandler(logWarningHandler{})
)

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)
	lh := logWarningHandler{logHandler{"got %q", []interface{}{"x"}}}
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

type noopHandlerT struct{}

func (noopHandlerT) Completed(time.Time) {}

var noopHandler = noopHandlerT{}

func deadlinedDeferred(t time.Duration) {
	defer Deadline(t, noopHandler).Done()
}

func deadlinedNotDeferred(t time.Duration) {
	s := Deadline(t, noopHandler)
	// Imagination
	s.Done()
}

func BenchmarkNoDurationDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedDeferred(0)
	}
}

func BenchmarkNoDurationNotDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedNotDeferred(0)
	}
}

func BenchmarkMSDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Deadline(time.Millisecond, nil).Done()
	}
}

func BenchmarkMSDurationDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedDeferred(time.Millisecond)
	}
}

func BenchmarkMSDurationNotDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedNotDeferred(time.Millisecond)
	}
}

func deadlinedLogDeferred(t time.Duration) {
	defer DeadlineLog(t, "").Done()
}

func deadlinedLogNotDeferred(t time.Duration) {
	s := DeadlineLog(t, "")
	// Imagination
	s.Done()
}

func deadlinedLogWarnDeferred(t time.Duration) {
	defer DeadlineLogWarn(t, "").Done()
}

func deadlinedLogWarnNotDeferred(t time.Duration) {
	s := DeadlineLogWarn(t, "")
	// Imagination
	s.Done()
}

func BenchmarkMSLogDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogDeferred(time.Millisecond)
	}
}

func BenchmarkMSLogNotDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogNotDeferred(time.Millisecond)
	}
}

func Benchmark0LogDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogDeferred(0)
	}
}

func Benchmark0LogNotDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogNotDeferred(0)
	}
}

func BenchmarkMSLogWarnDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogWarnDeferred(time.Millisecond)
	}
}

func BenchmarkMSLogWarnNotDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogWarnNotDeferred(time.Millisecond)
	}
}

func Benchmark0LogWarnDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogWarnDeferred(0)
	}
}

func Benchmark0LogWarnNotDeferred(b *testing.B) {
	for i := 0; i < b.N; i++ {
		deadlinedLogWarnNotDeferred(0)
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

func ExampleHandleFunc() {
	// Imagine a histogram tracking time spent in a particular
	// function.
	var histogram interface {
		Add(time.Duration)
	}

	// You can have a way to compute the time difference since a
	// particular point in time and then add it to your histogram:
	updateAgeHistogram := func(started time.Time) {
		d := time.Since(started)
		histogram.Add(d)
	}

	// And then you can use HandleFunc to invoke that handler:
	defer Deadline(time.Second, HandleFunc(updateAgeHistogram)).Done()

	// Do something that should take less than a second.  If it
	// takes more than a second, update the histogram with how
	// long it took.
}
