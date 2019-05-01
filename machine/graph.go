package machine

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	startID       = "start"
	endID         = "end"
	dotFileHeader = `
strict digraph stategraph {
	start [shape="circle", color="green", style="filled"]
	end [shape="circle", color="red", style="filled"]
`
	dotFileFooter = "}"
	dotEdgeStrFmt = "\t%s -> %s [label=\" %s\",fontsize=10]\n"
)

const (
	dotFileName      = "dot_graph"
	dotFileExtension = "gv"
)

// LogStateTransitionGraph enables tracing of states and transitions for the life of the finite-state machine.
// After the finite-state machine has reached a terminal or error state, a file named 'dot_graph.gv' will be
// created in the passed file path. The contents of the 'dot_graph.gv' file will be in DOT graph description
// language format, which allows rendering of a directed graph of the states and transitions made by the
// finite-state machine.
//
// If the passed string is empty, the file will be logged to the directory that the program was executed in.
func (fsm *FiniteStateMachine) LogStateTransitionGraph(path string) error {
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

// adjacencyMapToString writes the in-memory representation of the directed graph to a DOT formatted string.
func (fsm *FiniteStateMachine) adjacencyMapToString() error {
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
			stateGraphString := fmt.Sprintf(dotEdgeStrFmt, vertex, edge, stepLabel)

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
func (fsm *FiniteStateMachine) recordStateTransition(curr, next string) {
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

// writeStateGraphString ...
func (fsm *FiniteStateMachine) writeStateGraphString(str string) error {
	if fsm.dotFile != nil {
		_, err := fsm.dotFile.Write([]byte(str))
		if err != nil {
			return err
		}
	}
	return nil
}

// getFunctionName ...
func getFunctionName(f State) string {
	if f == nil {
		return endID
	}

	packageFuncName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	funcSegments := strings.Split(packageFuncName, "/")
	funcName := funcSegments[len(funcSegments)-1]
	return strings.Split(funcName, ".")[1]
}
