// Package fsm is a library that can be used to construct finite-state machines.
//
// By Austin Gebauer
package fsm

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	startID          = "start"
	endID            = "end"
	dotFileFooter    = "}"
	dotFileName      = "dot_graph"
	dotFileExtension = "gv"
	stepFontSize     = 10
	edgeFmtStr       = "\t%s -> %s [label=\" %s\",fontsize=%d]\n"
	dotFileHeader    = `
strict digraph stategraph {
	start [shape="circle", color="green", style="filled"]
	end [shape="circle", color="red", style="filled"]
`
)

// State is a function that handles a machine state and returns the next machine state.
//
// A function that participates as a state in the finite-state machine must be of the State type.
type State func() (State, error)

// A finiteStateMachine manages a finite-state machine.
type finiteStateMachine struct {
	start   State
	step    int64
	dotFile *os.File

	// adjacencyMap tracks each state and transition as a vertex to edge pair.
	// Each vertex to edge pair also records the step in which the transition happened.
	// Set(node) -> Set(edge) -> [step 1 ... step n]
	adjacencyMap map[string]map[string][]int64
}

// NewMachine initializes and returns a new finite-state machine.
func NewMachine() *finiteStateMachine {
	machine := finiteStateMachine{}
	machine.adjacencyMap = make(map[string]map[string][]int64)
	machine.step = 0
	return &machine
}

// Run starts the finite-state machine by invoking the passed State function.
// Run will continue to invoke State functions returned by State functions as
// the finite-state machine transitions from state to state.
// Run will return an error if an error is returned from any State function.
// Run will return nil if a terminal State is reached.
func (fsm *finiteStateMachine) Run(startState State) error {
	if startState == nil {
		return errors.New("start must not be nil")
	}

	fsm.start = startState
	err := fsm.run()

	if fsm.isTracing() {
		err := fsm.adjacencyMapToDotGraph()
		if err != nil {
			return err
		}
	}

	return err
}

// run starts the finite-state machine and records state transitions.
func (fsm *finiteStateMachine) run() error {
	var err error
	var currentState, nextState State
	currentState = fsm.start
	nextState = nil

	// Continue to process steps while not in a terminal state and an error hasn't occurred
	fsm.recordStateTransition(startID, getFunctionName(currentState))
	for currentState != nil && err == nil {
		nextState, err = currentState()
		fsm.recordStateTransition(getFunctionName(currentState), getFunctionName(nextState))
		currentState = nextState
	}

	return err
}

// LogStateTransitionGraph enables tracing of states and transitions for the life of the finite-state machine.
// After the finite-state machine has reached a terminal or error state, a file named 'dot_graph.gv' will be
// created in the passed file path. The contents of the 'dot_graph.gv' file will be in DOT graph description
// language format, which allows rendering of a directed graph of the states and transitions made by the
// finite-state machine.
//
// If the passed string is empty, the file will be logged to the directory that the program was executed in.
func (fsm *finiteStateMachine) LogStateTransitionGraph(path string) error {
	// If no path supplied, log to the directory that the program was executed in
	if path == "" {
		path = "."
	}
	filePath := fmt.Sprintf("%s/%s.%s", strings.TrimSuffix(path, "/"), dotFileName,
		dotFileExtension)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	fsm.dotFile = file
	return nil
}

// adjacencyMapToDotGraph writes the in-memory representation of the directed graph to a DOT formatted string.
func (fsm *finiteStateMachine) adjacencyMapToDotGraph() error {
	// write the header
	err := fsm.writeStateGraphString(dotFileHeader)
	if err != nil {
		return err
	}

	// write the graph state vertices, edges, and step counts
	for vertex, edges := range fsm.adjacencyMap {
		for edge, steps := range edges {
			stepBuf := bytes.Buffer{}
			for _, step := range steps {
				stepBuf.WriteString(fmt.Sprintf("%s,", strconv.FormatInt(step, 10)))
			}

			stepLabel := strings.TrimSuffix(stepBuf.String(), ",")
			stateGraphString := fmt.Sprintf(edgeFmtStr, vertex, edge, stepLabel, stepFontSize)

			err = fsm.writeStateGraphString(stateGraphString)
			if err != nil {
				return err
			}
		}
	}

	// write the footer
	err = fsm.writeStateGraphString(dotFileFooter)
	if err != nil {
		return err
	}

	return nil
}

// recordStateTransition records a state transition in the finite-state machine.
func (fsm *finiteStateMachine) recordStateTransition(curr, next string) {
	if !fsm.isTracing() {
		return
	}

	// Increase the step count
	fsm.step++

	// Add state vertex and edge
	_, haveVertex := fsm.adjacencyMap[curr]

	// Create a new state vertex if it does not already exist
	if !haveVertex {
		fsm.adjacencyMap[curr] = map[string][]int64{}
	}
	edgeMap, _ := fsm.adjacencyMap[curr]

	// Create a new edge to the state vertex if it does not already exist
	_, haveEdge := edgeMap[next]
	if !haveEdge {
		edgeMap[next] = []int64{}
	}
	edgeSteps, _ := edgeMap[next]

	// Append the step count into the edge steps
	edgeMap[next] = append(edgeSteps, fsm.step)
}

// writeStateGraphString writes the passed string into the dot file if it exists.
func (fsm *finiteStateMachine) writeStateGraphString(str string) error {
	if fsm.dotFile != nil {
		_, err := fsm.dotFile.Write([]byte(str))
		if err != nil {
			return err
		}
	}
	return nil
}

// isTracing returns true if the finite-state machine has been configured to trace states and transitions.
func (fsm *finiteStateMachine) isTracing() bool {
	return fsm.dotFile != nil
}

// getFunctionName returns the name of the passed State function.
// The package name in which the function exists will be stripped from the returned name.
func getFunctionName(f State) string {
	if f == nil {
		return endID
	}

	packageFuncName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	funcSegments := strings.Split(packageFuncName, "/")
	funcName := funcSegments[len(funcSegments)-1]
	return strings.Split(funcName, ".")[1]
}
