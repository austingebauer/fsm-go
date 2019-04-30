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
// A function that participates as a state in the finite-state machine must by of the State type.
type State func() (State, error)

// NewMachine initializes and returns a new finite-state machine.
func NewMachine() *FiniteStateMachine {
	machine := FiniteStateMachine{}
	machine.adjacencyMap = make(map[string]map[string][]int64)
	machine.step = 0
	return &machine
}

// Run starts the finite-state machine by invoking the passed State function.
// Run will continue to invoke State functions returned by State functions as
// the finite-state machine transitions from state to state.
// Run will return an error if an error is returned from any State function.
// Run will return nil if a terminal State is reached.
func (fsm *FiniteStateMachine) Run(startState State) error {
	if startState == nil {
		return errors.New("start must not be nil")
	}

	fsm.start = startState
	err := fsm.run()

	if fsm.isTracing() {
		err := fsm.adjacencyMapToString()
		if err != nil {
			return err
		}
	}

	return err
}

// run starts the finite-state machine and records state transitions.
func (fsm *FiniteStateMachine) run() error {
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

// isTracing returns true if the finite-state machine has been configured to trace states and transitions.
func (fsm *FiniteStateMachine) isTracing() bool {
	return fsm.dotFile != nil
}
