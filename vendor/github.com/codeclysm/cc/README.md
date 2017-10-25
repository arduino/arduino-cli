# cc
A golang waitgroup with limits and error reporting

Usage:

```go
import github.com/codeclysm/cc

p := cc.New(4)
p.Run(func() error {
	return errors.New("fail1")
})
p.Run(func() error {
	return errors.New("fail2")
})
p.Run(func() error {
	return nil
})

errs := p.Wait() // returns a list of errors [fail1, fail2]
```