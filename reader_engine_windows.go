package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// The reader serializes requests. The lock coordinates process startup and
// shutdown with the UI goroutine without holding it during rendering.
type readerEngine struct {
	mu         sync.Mutex
	process    *readerProcess
	closed     bool
	background bool
}

type readerProcess struct {
	command *exec.Cmd
	input   io.WriteCloser
	output  io.ReadCloser
	send    *json.Encoder
	receive *json.Decoder
}

func (e *readerEngine) start() (*readerProcess, error) {
	e.mu.Lock()
	if e.process != nil || e.closed {
		process, closed := e.process, e.closed
		e.mu.Unlock()
		if closed {
			return nil, os.ErrClosed
		}
		return process, nil
	}
	e.mu.Unlock()

	cmd, err := engineCommand("--stream")
	if err != nil {
		return nil, err
	}
	if e.background {
		cmd.SysProcAttr.CreationFlags |= 0x00004000 // BELOW_NORMAL_PRIORITY_CLASS
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil, os.ErrClosed
	}
	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		input.Close()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("Zpracování náhledu se nepodařilo spustit.\n%w", err)
	}
	e.process = &readerProcess{cmd, input, output, json.NewEncoder(input), json.NewDecoder(output)}
	return e.process, nil
}

func (e *readerEngine) run(req request) response {
	process, err := e.start()
	if err != nil {
		return operationError(err.Error())
	}
	var result response
	if process.send.Encode(req) == nil && process.receive.Decode(&result) == nil && result.Status != "" {
		return result
	}
	e.mu.Lock()
	retry := e.process == process
	if retry {
		e.process = nil
	}
	e.mu.Unlock()
	if !retry {
		return operationError("Načítání náhledu bylo ukončeno.")
	}
	process.command.Process.Kill()
	process.command.Wait()
	if e.background {
		return operationError("Přednačtení stránky se nezdařilo.")
	}
	// A failed exchange can be retried because preview requests only read the PDF.
	return runOperation(req)
}

func (e *readerEngine) close() {
	e.mu.Lock()
	process := e.process
	e.process, e.closed = nil, true
	e.mu.Unlock()
	if process != nil {
		// EOF lets Python clean up any active render before it exits.
		process.input.Close()
		process.output.Close()
		go process.command.Wait()
	}
}
