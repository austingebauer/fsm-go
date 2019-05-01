# go-fsm

A Golang library that can be used to construct finite-state machines.

A finite-state machine is an abstract machine that can be in exactly one of a finite number of states at any 
given time.

## Installation

```bash
go get github.com/austingebauer/go-fsm
```

## Usage

See usage example in `example/main.go`, which uses go-fsm to write a finite-state machine for the 
[pacman ghosts state machine](https://bits.theorem.co/images/posts/2015-01-21-state-design-pacman-fsm.png).

To run the example:
```bash
go run example/main.go
```

To view the graph diagram for the life of the finite-state machine:
```bash
dot example/dot_graph.gv -T png | open -f -a /Applications/Preview.app
```
