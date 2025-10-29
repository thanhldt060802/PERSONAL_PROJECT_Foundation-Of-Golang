package app

import (
	"fmt"
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// State
const (
	Idle          = "idle"
	GettingNumber = "getting_number"
	RingingASide  = "ringing_a_side"
	RingingBSide  = "ringing_b_side"
	Speech        = "speech"
	WaitOnHook    = "wait_on_hook"
)

// Event
const (
	AOffHook                 = "a_off_hook"
	AOnHook                  = "a_on_hook"
	DialledNoBusyOrIncorrect = "dialled_no_busy_or_incorrect"
	BSideAcceptsCall         = "b_side_accepts_call"
	CallFromOtherTelephone   = "call_from_other_telephone"
	BSideAnswer              = "b_side_answer"
	OtherSideGoesOnHook      = "other_side_goes_on_hook"
	OwnSideGoesOnHook        = "own_side_goes_on_hook"
)

type ReceiverFSMActor struct {
	act.Actor

	state string
}

func FactoryReceiverFSMActor() gen.ProcessBehavior {
	return &ReceiverFSMActor{}
}

func (receiverFSMActor *ReceiverFSMActor) Init(args ...any) error {
	receiverFSMActor.state = Idle
	receiverFSMActor.Log().Info("Started process %s %s on %s", receiverFSMActor.PID(), receiverFSMActor.Name(), receiverFSMActor.Node().Name())

	nextEvent := []string{AOffHook, CallFromOtherTelephone}[rand.Intn(2)]
	receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)

	return nil
}

func (receiverFSMActor *ReceiverFSMActor) HandleMessage(from gen.PID, message any) error {
	switch receiverFSMActor.state {
	case Idle:
		{
			receiverFSMActor.Idle(message)
			return nil
		}
	case GettingNumber:
		{
			receiverFSMActor.GettingNumber(message)
			return nil
		}
	case RingingASide:
		{
			receiverFSMActor.RingingASide(message)
			return nil
		}
	case RingingBSide:
		{
			receiverFSMActor.RingingBSide(message)
			return nil
		}
	case Speech:
		{
			receiverFSMActor.Speech(message)
			return nil
		}
	case WaitOnHook:
		{
			receiverFSMActor.WaitOnHook(message)
			return nil
		}
	}

	return nil
}

func (receiverFSMActor *ReceiverFSMActor) Idle(message any) {
	fmt.Println()
	event := message.(string)
	receiverFSMActor.Log().Info("--- With event: %s, current state: %s", event, receiverFSMActor.state)
	switch event {
	case AOffHook:
		{
			receiverFSMActor.state = GettingNumber
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOnHook, BSideAcceptsCall, DialledNoBusyOrIncorrect}[rand.Intn(3)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	case CallFromOtherTelephone:
		{
			receiverFSMActor.state = RingingBSide
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOnHook, BSideAnswer}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	}
}

func (receiverFSMActor *ReceiverFSMActor) GettingNumber(message any) {
	event := message.(string)
	receiverFSMActor.Log().Info("--- With event: %s, current state: %s", event, receiverFSMActor.state)
	switch event {
	case DialledNoBusyOrIncorrect:
		{
			receiverFSMActor.state = WaitOnHook
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), OwnSideGoesOnHook, 2*time.Second)
			return
		}
	case BSideAcceptsCall:
		{
			receiverFSMActor.state = RingingASide
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOnHook, BSideAnswer}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	case AOnHook:
		{
			receiverFSMActor.state = Idle
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOffHook, CallFromOtherTelephone}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	}
}

func (receiverFSMActor *ReceiverFSMActor) RingingASide(message any) {
	event := message.(string)
	receiverFSMActor.Log().Info("--- With event: %s, current state: %s", event, receiverFSMActor.state)
	switch event {
	case AOnHook:
		{
			receiverFSMActor.state = Idle
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOffHook, CallFromOtherTelephone}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	case BSideAnswer:
		{
			receiverFSMActor.state = Speech
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{OwnSideGoesOnHook, OtherSideGoesOnHook}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	}
}

func (receiverFSMActor *ReceiverFSMActor) RingingBSide(message any) {
	event := message.(string)
	receiverFSMActor.Log().Info("--- With event: %s, current state: %s", event, receiverFSMActor.state)
	switch event {
	case AOnHook:
		{
			receiverFSMActor.state = Idle
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOffHook, CallFromOtherTelephone}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	case BSideAnswer:
		{
			receiverFSMActor.state = Speech
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{OwnSideGoesOnHook, OtherSideGoesOnHook}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	}
}

func (receiverFSMActor *ReceiverFSMActor) Speech(message any) {
	event := message.(string)
	receiverFSMActor.Log().Info("--- With event: %s, current state: %s", event, receiverFSMActor.state)
	switch event {
	case OwnSideGoesOnHook:
		{
			receiverFSMActor.state = Idle
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOffHook, CallFromOtherTelephone}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	case OtherSideGoesOnHook:
		{
			receiverFSMActor.state = WaitOnHook
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), OwnSideGoesOnHook, 2*time.Second)
			return
		}
	}
}

func (receiverFSMActor *ReceiverFSMActor) WaitOnHook(message any) {
	event := message.(string)
	receiverFSMActor.Log().Info("--- With event: %s, current state: %s", event, receiverFSMActor.state)
	switch event {
	case OwnSideGoesOnHook:
		{
			receiverFSMActor.state = Idle
			receiverFSMActor.Log().Info("--> Transition to %s", receiverFSMActor.state)
			nextEvent := []string{AOffHook, CallFromOtherTelephone}[rand.Intn(2)]
			receiverFSMActor.SendAfter(receiverFSMActor.PID(), nextEvent, 2*time.Second)
			return
		}
	}
}
