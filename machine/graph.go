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
)

// LogStateTransitionGraph ...
func (sm *FiniteStateMachine) LogStateTransitionGraph(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	sm.dotFile = file
	return nil
}

// writeStateGraph ...
func (sm *FiniteStateMachine) writeStateGraph() error {
	// write the header
	err := sm.writeStateGraphString(dotFileHeader)
	if err != nil {
		return err
	}

	// write the graph state verticies, edges, and step counts
	for vertex, edges := range sm.adjacencyMap {
		for edge, steps := range edges {
			stepBuf := bytes.Buffer{}
			for _, step := range steps {
				stepBuf.WriteString(fmt.Sprintf("%s,", strconv.FormatInt(step, 10)))
			}

			stepLabel := strings.TrimSuffix(stepBuf.String(), ",")
			stateGraphString := fmt.Sprintf("\t%s -> %s [label=\"%s\"]\n", vertex, edge, stepLabel)

			err = sm.writeStateGraphString(stateGraphString)
			if err != nil {
				return err
			}
		}
	}

	// write the footer
	err = sm.writeStateGraphString(dotFileFooter)
	if err != nil {
		return err
	}

	return nil
}

// recordStateTransition records a state transition in the finite-state machine.
func (sm *FiniteStateMachine) recordStateTransition(curr, next string) {
	// Increase the step count
	sm.step++

	// Add state vertex and edge
	_, haveVertex := sm.adjacencyMap[curr]

	// Create a new state vertex if it does not already exist
	if !haveVertex {
		sm.adjacencyMap[curr] = map[string][]int64{}
	}
	edgeMap, _ := sm.adjacencyMap[curr]

	// Create a new edge to the state vertex if it does not already exist
	_, haveEdge := edgeMap[next]
	if !haveEdge {
		edgeMap[next] = []int64{}
	}
	edgeSteps, _ := edgeMap[next]

	// Append the step count into the edge steps
	edgeMap[next] = append(edgeSteps, sm.step)
}

// writeStateGraphString ...
func (sm *FiniteStateMachine) writeStateGraphString(str string) error {
	if sm.dotFile != nil {
		_, err := sm.dotFile.Write([]byte(str))
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
