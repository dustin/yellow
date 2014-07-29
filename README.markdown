Yellow helps raise awareness of functions that are taking longer than
expected.

In the most simple cases, you can log slow tasks (with whatever
context you want to supply), but you can also react to slowness by
having code perform some task on timeout.  See
[the documentation][docs] for details and examples.

## Example

Basic usage (send slow invocations to the standard logger):

```go
func ExampleDeadlineLog() {
	defer DeadlineLog(time.Second, "Doing thing %d", 1).Done()
	// do something that should take less than a second, log
	// otherwise
}
```

[![Coverage Status](https://coveralls.io/repos/dustin/yellow/badge.png?branch=master)](https://coveralls.io/r/dustin/yellow?branch=master)

[docs]: http://godoc.org/github.com/dustin/yellow
