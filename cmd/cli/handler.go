package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cristalhq/acmd"
	"github.com/google/shlex"
)

type InputHandler interface {
	HandleInput()
}

type inputHandler struct {
	in       *bufio.Reader
	out      io.Writer
	ctx      context.Context
	cancel   context.CancelFunc
	commands []acmd.Command
}

func NewHandler(ctx context.Context, cancel context.CancelFunc, in io.ReadCloser, out io.Writer,
	commands []acmd.Command,
) InputHandler {
	return &inputHandler{
		in:       bufio.NewReader(in),
		out:      out,
		ctx:      ctx,
		cancel:   cancel,
		commands: commands,
	}
}

func (h *inputHandler) HandleInput() {
	defer h.cancel()

	var err error
	var input string

	if !h.writeStr("Connected to " + getConnStr()) {
		return
	}

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			input, err = h.in.ReadString(byte('\n'))
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				if !h.writeStr(err.Error()) {
					return
				}
			}
			input = strings.TrimSpace(input)

			args, err := shlex.Split(input)
			if err != nil {
				if !h.writeStr(err.Error()) {
					return
				}
			}

			r := acmd.RunnerOf(
				h.commands, acmd.Config{
					AppName:         "",
					Version:         "1.0",
					AppDescription:  "CLI for anti-bruteforce service",
					PostDescription: "",
					Output:          h.out,
					Args:            append([]string{""}, args...),
					Context:         h.ctx,
					Usage:           usage,
				},
			)

			if err = r.Run(); err != nil {
				if !errors.Is(err, acmd.ErrNoArgs) && !strings.Contains(err.Error(), "no such command") {
					if !h.writeStr("Error: " + err.Error()) {
						return
					}
				}
			}
		}
	}
}

func (h *inputHandler) writeStr(str string) bool {
	_, err := h.out.Write([]byte(str + "\n"))
	if err != nil {
		printStderr(err.Error())
		return false
	}
	return true
}
