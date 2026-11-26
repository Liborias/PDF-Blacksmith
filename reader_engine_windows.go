package main

import (
	"bufio"
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
	receive *bufio.Reader
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

	cmd, err := engineCommand("--reader-stream")
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
	e.process = &readerProcess{cmd, input, output, json.NewEncoder(input), bufio.NewReader(output)}
	return e.process, nil
}

func (p *readerProcess) readResponse(action string) (response, error) {
	var result response
	header, err := p.receive.ReadBytes('\n')
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(header, &result); err != nil {
		return result, err
	}
	view := result.View
	if view != nil && (view.RasterWidth != 0 || view.RasterHeight != 0) {
		// Screen frames are packed, top-down BGRX pixels after the JSON line.
		w, h := view.RasterWidth, view.RasterHeight
		if w < 1 || h < 1 || w > screenRegionEdge || h > screenRegionEdge || w*h > 3200*3200 {
			return result, fmt.Errorf("neplatné rozměry náhledu %d × %d", w, h)
		}
		view.Image = make([]byte, w*h*4)
		if _, err := io.ReadFull(p.receive, view.Image); err != nil {
			return result, err
		}
		if view.DragLayer {
			if action != "object_drag_preview" || w*h > 2_000_000 {
				return result, fmt.Errorf("neplatná vrstva náhledu přesunu")
			}
			view.dragPixels = make([]byte, len(view.Image))
			if _, err := io.ReadFull(p.receive, view.dragPixels); err != nil {
				return result, err
			}
			// Win32 AlphaBlend expects premultiplied BGRA.
			for i := 0; i < len(view.dragPixels); i += 4 {
				alpha := uint16(view.dragPixels[i+3])
				for j := range 3 {
					view.dragPixels[i+j] = byte((uint16(view.dragPixels[i+j])*alpha + 127) / 255)
				}
			}
		}
	}
	return result, nil
}

func (e *readerEngine) run(req request) response {
	process, err := e.start()
	if err != nil {
		return operationError(err.Error())
	}
	if process.send.Encode(req) == nil {
		result, err := process.readResponse(req.Action)
		if err == nil && result.Status != "" {
			return result
		}
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

func (e *readerEngine) close() <-chan struct{} {
	e.mu.Lock()
	process := e.process
	e.process, e.closed = nil, true
	e.mu.Unlock()
	if process == nil {
		return nil
	}
	done := make(chan struct{})
	// Terminate off the UI thread before releasing synchronous Windows pipes.
	go func() {
		defer close(done)
		process.command.Process.Kill()
		process.input.Close()
		process.output.Close()
		process.command.Wait()
	}()
	return done
}
