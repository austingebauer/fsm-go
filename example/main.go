package main

import (
	"errors"
	"fmt"
	"github.com/austingebauer/go-fsm"
	"log"
	"math/rand"
	"time"
)

func main() {
	// Initialize a new machine
	sm := fsm.NewMachine()

	// (optional) Log state transitions to a DOT graph description language file
	err := sm.LogStateTransitionGraph("./example")
	if err != nil {
		log.Fatal(err)
	}

	// Run the finite-state machine until an error occurs or a terminal state is reached
	err := sm.Run(WanderState)
	if err != nil {
		// An error occurred in a state and was returned from a state
		log.Fatal(err)
	}

	// Terminal state reached
	fmt.Println("game over: ghost ate pacman")
}

// The handler for pacman ghost WANDER_STATE.
func WanderState() (fsm.State, error) {
	if randInt()%2 == 0 {
		// Spotted pacman. Chase.
		return ChaseState, nil
	} else {
		// Pacman has ate the power pellet. Flee.
		return FleeState, nil
	}
}

// The handler for pacman ghost CHASE_STATE.
func ChaseState() (fsm.State, error) {
	if randInt()%2 == 0 {
		// Lose or eat pacman
		if randInt()%2 == 0 {
			// Ate pacman. Game over.
			return nil, nil
		} else {
			// Lost pacman. Wander.
			return WanderState, nil
		}
	} else {
		if randInt()%2 == 0 {
			// An error occured. Pacman glitch..
			return nil, errors.New("error: pacman glitch")
			// return FleeState, nil
		} else {
			// Pacman has ate the power pellet. Flee.
			return FleeState, nil
		}
	}
}

// The handler for pacman ghost RETURN_TO_BASE_STATE.
func ReturnToBaseState() (fsm.State, error) {
	// Reached central base. Start wandering again.
	return WanderState, nil
}

// The handler for pacman ghost FLEE_STATE.
func FleeState() (fsm.State, error) {
	if randInt()%2 == 0 {
		// Power pellet expires. Wander.
		return WanderState, nil
	} else {
		// Eaten by pacman. Return to base.
		return ReturnToBaseState, nil
	}
}

func randInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}
