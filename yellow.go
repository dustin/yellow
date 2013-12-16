// Package yellow helps you report on abnormally slow functions in
// production applications.
//
// Example:
//
//   func ShouldBeFast(thing, place string) {
//       defer yellow.DeadlineLog(time.Second,
//           "getting %q from %v", thing, place).Done()
//       doThing(thing, place)
//   }
//
package yellow

import (
	"log"
	"time"
)

// Handler handles deadline events.
type Handler interface {
	// TimedOut is called when your Deadline has exceeded.
	TimedOut(started time.Time)
	// Completed is called when a task that has exceeded its
	// deadline finally completes.
	Completed(started time.Time)
}

// Stopwatch manages a timer that runs while waiting for a deadline.
type Stopwatch struct {
	handler Handler
	started time.Time
	t       *time.Timer
}

// Done allows the caller to indicate the Deadlined function has completed.
func (d *Stopwatch) Done() {
	if d == nil {
		return
	}
	if !d.t.Stop() {
		d.handler.Completed(d.started)
	}
}

// Deadline sets up a Handler to be notified if Done isn't called
// before the requested timeout occurs.
func Deadline(d time.Duration, handler Handler) *Stopwatch {
	if d == 0 {
		return nil
	}
	started := time.Now()
	return &Stopwatch{
		handler, started,
		time.AfterFunc(d, func() { handler.TimedOut(started) }),
	}
}

// DeadlineLog is a convenience invocation of Deadline that just logs events.
func DeadlineLog(d time.Duration, format string, args ...interface{}) *Stopwatch {
	return Deadline(d, logHandler{format, args})
}

// LogHandler is a handler that logs handled events.
type logHandler struct {
	// format string passed to the underlying logger. The duration will be appended to this.
	format string
	// args for the format string
	args []interface{}
}

// TimedOut satisfies Handler.Timeout
func (l logHandler) TimedOut(started time.Time) {
	log.Printf("Taking too long: "+l.format+" "+time.Since(started).String(), l.args...)
}

// Completed satisfies Handler.Completed
func (l logHandler) Completed(started time.Time) {
	log.Printf("Finally finished: "+l.format+" "+time.Since(started).String(), l.args...)
}
