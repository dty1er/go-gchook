## go-gchook



Inject arbitary actions when go garbage collector worked

## Installation

```
$ go get -u github.com/yagi5/go-gchook
```

## Usage

You can also try this with [The Go Playground](https://play.golang.org/p/4IPSzsfkSuL).

```
package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/yagi5/go-gchook"
)

func main() {
	done := make(chan struct{}, 1)
	memstats := &runtime.MemStats{}
	numGC := uint32(0)
	gchook.Register(func() {
		runtime.ReadMemStats(memstats)
		numGC++
		if numGC != memstats.NumGC {
			log.Fatal("Skipped a GC notification")
		}
		if numGC >= 100 {
			gchook.Cancel()
			done <- struct{}{}
		}
	})
LOOP:
	for {
		select {
		case <-time.After(1 * time.Millisecond):
			// copied from: https://golang.org/test/gc.go
			gc1()
			runtime.GC()
		case <-done:
			break LOOP
		}
	}
	fmt.Printf("numGC = %+v\n", numGC)
	fmt.Printf("memstats.NumGC = %+v\n", memstats.NumGC)
}

func gc1() {
	gc2()
}

func gc2() {
	b := new([10000]byte)
	_ = b
}
```

## Author

[Hidetatsu Yaginuma](https://github.com/yagi5)

## LICENSE

[MIT](https://github.com/yagi5/go-gchook/blob/master/LICENSE)
