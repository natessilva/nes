# NES

A partially complete NES emulator written in Go. Compiles to WASM and available for play at https://nesdev.netlify.app.

# Summary

This emulator exposes a relatively simple Go package with a minimal API that can be wrapped with pretty much any GUI. The currently implemented wrapper is a WASM website, but one could easily also build a native GUI on top of this package with minimal effort.

I chose WASM because I wanted to learn about the tech and see how good performance would be. On my local setup, performance playing titles like Pac-Man, Donkey Kond and Super Mario Bros appears to be acceptabe.

# Building

The WASM website is relatively naive, just an index.html the Go supplied wasm_exec.js and the built nes.wasm file.

Simply compile cmd/wasm/main.go and host the contents on web/ online!

```
GOOS=js GOARCH=wasm go build -o web/nes.wasm cmd/wasm/main.go
```

# Testing

Testing of the CPU is based on the [golden log](https://www.qmtpro.com/~nes/misc/nestest.log) file. It loads nestes.nes and steps through the log file verifying cycle counts and register values.

```
go test
```

# Usage

The website can be run locally by running

```
go run cmd/server/main.go
```

And loading http://localhost:8000.

The website is super basic as of yet. It consists of a file input and a canvas. Load your favorite (legally obtained of course 😉) ROM and begin playing!

# Controls

Controls are hardcoded and only keyboard controls are supported currently.

Arrow keys, Enter, Space, D and F map to the NES arrow buttons, Start, Select, B, and A buttons respectively

# Frame timing

60 FPS is acheived with requestAnimationFrame and some naive logic to skip frames when that is too fast. It appears to work well on my local machine for a time, but also appears to slow down periodically. I imagine this might be challenging to overcome in a javascript environment.

# Work remaining

In short, a lot 😅

- NROM is the only mapper currently implemented
- Audio support has not been started
- Put more design effort into the website
