// This file was generated by counterfeiter
package botfakes

import (
	"sync"

	"github.com/mdelillo/claimer/bot/commands"
)

type FakeCommandFactory struct {
	NewCommandStub        func(command string, args []string, username string) commands.Command
	newCommandMutex       sync.RWMutex
	newCommandArgsForCall []struct {
		command  string
		args     []string
		username string
	}
	newCommandReturns struct {
		result1 commands.Command
	}
	newCommandReturnsOnCall map[int]struct {
		result1 commands.Command
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCommandFactory) NewCommand(command string, args []string, username string) commands.Command {
	var argsCopy []string
	if args != nil {
		argsCopy = make([]string, len(args))
		copy(argsCopy, args)
	}
	fake.newCommandMutex.Lock()
	ret, specificReturn := fake.newCommandReturnsOnCall[len(fake.newCommandArgsForCall)]
	fake.newCommandArgsForCall = append(fake.newCommandArgsForCall, struct {
		command  string
		args     []string
		username string
	}{command, argsCopy, username})
	fake.recordInvocation("NewCommand", []interface{}{command, argsCopy, username})
	fake.newCommandMutex.Unlock()
	if fake.NewCommandStub != nil {
		return fake.NewCommandStub(command, args, username)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.newCommandReturns.result1
}

func (fake *FakeCommandFactory) NewCommandCallCount() int {
	fake.newCommandMutex.RLock()
	defer fake.newCommandMutex.RUnlock()
	return len(fake.newCommandArgsForCall)
}

func (fake *FakeCommandFactory) NewCommandArgsForCall(i int) (string, []string, string) {
	fake.newCommandMutex.RLock()
	defer fake.newCommandMutex.RUnlock()
	return fake.newCommandArgsForCall[i].command, fake.newCommandArgsForCall[i].args, fake.newCommandArgsForCall[i].username
}

func (fake *FakeCommandFactory) NewCommandReturns(result1 commands.Command) {
	fake.NewCommandStub = nil
	fake.newCommandReturns = struct {
		result1 commands.Command
	}{result1}
}

func (fake *FakeCommandFactory) NewCommandReturnsOnCall(i int, result1 commands.Command) {
	fake.NewCommandStub = nil
	if fake.newCommandReturnsOnCall == nil {
		fake.newCommandReturnsOnCall = make(map[int]struct {
			result1 commands.Command
		})
	}
	fake.newCommandReturnsOnCall[i] = struct {
		result1 commands.Command
	}{result1}
}

func (fake *FakeCommandFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newCommandMutex.RLock()
	defer fake.newCommandMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCommandFactory) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
