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
// If your handler also implements TimedOutHandler,
// TimedOut(time.Time) will be delivered to your Handler while the
// function is still running.
//
// This is useful, for example, to deliver warnings about functions
// that are running so slowly as to be completely unresponsive.
package yellow

import (
	"log"
	"time"
)

// Handler receives notifications when tasks complete after exceeding
// their deadlines.
type Handler interface {
	// Completed is called when a task that has exceeded its
	// deadline finally completes.
	Completed(started time.Time)
}

type funcHandler func(time.Time)

// Completed satisfies Handler
func (f funcHandler) Completed(t time.Time) {
	f(t)
}

// HandleFunc wraps a simple function to satisfy Handler
func HandleFunc(f func(t time.Time)) Handler {
	return funcHandler(f)
}

// TimedOutHandler receives notifications as Deadlines are exceeded,
// but while the task is still running.
type TimedOutHandler interface {
	Handler
	// TimedOut is called when your Deadline has exceeded.
	TimedOut(started time.Time)
}

// Stopwatch manages a timer that runs while waiting for a deadline.
type Stopwatch struct {
	handler Handler
	started time.Time
	d       time.Duration
	t       *time.Timer
}

// Done allows the caller to indicate the Deadlined function has completed.
func (d *Stopwatch) Done() {
	if d == nil {
		return
	}
	if d.t == nil {
		if time.Since(d.started) > d.d {
			d.handler.Completed(d.started)
		}
	} else {
		if !d.t.Stop() {
			d.handler.Completed(d.started)
		}
	}
}

// Deadline sets up a Handler to be notified if Done isn't called
// before the requested timeout occurs.
func Deadline(d time.Duration, handler Handler) *Stopwatch {
	if d == 0 {
		return nil
	}
	started := time.Now()
	rv := &Stopwatch{handler, started, d, nil}
	if h, ok := handler.(TimedOutHandler); ok {
		rv.t = time.AfterFunc(d, func() { h.TimedOut(started) })
	}
	return rv
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
