package machine

import (
	"github.com/pkg/errors"
	"os"
)

// A FiniteStateMachine manages a finite-state machine.
type FiniteStateMachine struct {
	start   State
	step    int64
	dotFile *os.File

	// adjacencyMap tracks each state and transition as a vertex to edge pair.
	// Each vertex to edge pair also records the step in which the transition happened.
	// Set(node) -> Set(edge) -> [step 1 ... step n]
	adjacencyMap map[string]map[string][]int64
}

// State is a function that handles a machine state and returns the next machine state.
type State func() (State, error)

// NewMachine initializes and returns a new finite-state machine.
func NewMachine() *FiniteStateMachine {
	machine := FiniteStateMachine{}
	machine.adjacencyMap = make(map[string]map[string][]int64)
	machine.step = 0
	return &machine
}

// Run starts the finite-state machine by invoking the passed State.
func (sm *FiniteStateMachine) Run(startState State) error {
	if startState == nil {
		return errors.New("start must not be nil")
	}

	sm.start = startState
	err := sm.run()

	if sm.dotFile != nil {
		err := sm.writeStateGraph()
		if err != nil {
			return err
		}
	}

	return err
}

// run ...
func (sm *FiniteStateMachine) run() error {
	var err error
	var currentState, nextState State
	currentState = sm.start
	nextState = nil

	// Continue to process steps while not in a terminal state and an error hasn't occurred
	sm.recordStateTransition(startID, getFunctionName(currentState))
	for currentState != nil && err == nil {
		nextState, err = currentState()
		sm.recordStateTransition(getFunctionName(currentState), getFunctionName(nextState))
		currentState = nextState
	}

	return err
}
