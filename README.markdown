Yellow helps raise awareness of functions that are taking longer than
expected.

## Example

Basic usage (send slow invocations to the standard logger):

```go
func ExampleDeadlineLog() {
	defer DeadlineLog(time.Second, "Doing thing %d", 1).Done()
	// do something that should take less than a second, log
	// otherwise
}
```
