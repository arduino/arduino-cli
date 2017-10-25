bonjour
====

This is a simple Multicast DNS-SD (Apple Bonjour) library written in Golang. You can use it to discover services in the LAN. Pay attention to the infrastructure you are planning to use it (clouds or shared infrastructures usually prevent mDNS from functioning). But it should work in the most office, home and private environments.

**IMPORTANT**: It does NOT pretend to be a full & valid implementation of the RFC 6762 & RFC 6763, but it fulfils the requirements of its authors (we just needed service discovery in the LAN environment for our IoT products). The registration code needs a lot of improvements. This code was not tested for Bonjour conformance but have been manually verified to be working using built-in OSX utility `/usr/bin/dns-sd`.

Detailed documentation: [![GoDoc](https://godoc.org/github.com/oleksandr/bonjour?status.svg)](https://godoc.org/github.com/oleksandr/bonjour)


##Browsing available services in your local network

Here is an example how to browse services by their type:

```
package main

import (
    "log"
    "os"
    "time"

    "github.com/oleksandr/bonjour"
)

func main() {
    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Println("Failed to initialize resolver:", err.Error())
        os.Exit(1)
    }

    results := make(chan *bonjour.ServiceEntry)

    go func(results chan *bonjour.ServiceEntry, exitCh chan<- bool) {
        for e := range results {
            log.Printf("%s", e.Instance)
            exitCh <- true
            time.Sleep(1e9)
            os.Exit(0)
        }
    }(results, resolver.Exit)

    err = resolver.Browse("_foobar._tcp", "local.", results)
    if err != nil {
        log.Println("Failed to browse:", err.Error())
    }

    select {}
}
```

##Doing a lookup of a specific service instance

Here is an example of looking up service by service instance name:

```
package main

import (
    "log"
    "os"
    "time"

    "github.com/oleksandr/bonjour"
)

func main() {
    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Println("Failed to initialize resolver:", err.Error())
        os.Exit(1)
    }

    results := make(chan *bonjour.ServiceEntry)

    go func(results chan *bonjour.ServiceEntry, exitCh chan<- bool) {
        for e := range results {
            log.Printf("%s", e.Instance)
            exitCh <- true
            time.Sleep(1e9)
            os.Exit(0)
        }
    }(results, resolver.Exit)

    err = resolver.Lookup("DEMO", "_foobar._tcp", "", results)
    if err != nil {
        log.Println("Failed to browse:", err.Error())
    }

    select {}
}
```


##Registering a service

Registering a service is as simple as the following:

```
package main

import (
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/oleksandr/bonjour"
)

func main() {
    // Run registration (blocking call)
    s, err := bonjour.Register("Foo Service", "_foobar._tcp", "", 9999, []string{"txtv=1", "app=test"}, nil)
    if err != nil {
        log.Fatalln(err.Error())
    }

    // Ctrl+C handling
    handler := make(chan os.Signal, 1)
    signal.Notify(handler, os.Interrupt)
    for sig := range handler {
        if sig == os.Interrupt {
            s.Shutdown()
            time.Sleep(1e9)
            break
        }
    }
}
```


##Registering a service proxy (manually specifying host/ip and avoiding lookups)

```
package main

import (
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/oleksandr/bonjour"
)

func main() {
    // Run registration (blocking call)
    s, err := bonjour.RegisterProxy("Proxy Service", "_foobar._tcp", "", 9999, "octopus", "10.0.0.111", []string{"txtv=1", "app=test"}, nil)
    if err != nil {
        log.Fatalln(err.Error())
    }

    // Ctrl+C handling
    handler := make(chan os.Signal, 1)
    signal.Notify(handler, os.Interrupt)
    for sig := range handler {
        if sig == os.Interrupt {
            s.Shutdown()
            time.Sleep(1e9)
            break
        }
    }
}
```