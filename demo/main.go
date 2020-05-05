// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"github.com/mzyy94/nscon"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func main() {
	target := "/dev/hidg0"
	name := "procon"
	con := nscon.NewController(target, name)
	con.LogLevel = 2
	defer con.Close()
	con.Connect()

	buf := make([]byte, 1)

	// Set tty break for read keyboard input directly
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			con.Close()
			os.Exit(130)
		}
	}()

	for {
		os.Stdin.Read(buf)
		switch buf[0] {
		case 'a':
			con.Input.Dpad.Left = 1
		case 'd':
			con.Input.Dpad.Right = 1
		case 'w':
			con.Input.Dpad.Up = 1
		case 's':
			con.Input.Dpad.Down = 1
		case ' ':
			con.Input.Button.B = 1
		case 0x0a: // Enter
			con.Input.Button.A = 1
		case '.':
			con.Input.Button.X = 1
		case '/':
			con.Input.Button.Y = 1
		case 0x1b: // Escape
			con.Input.Button.Home = 1
		case '`':
			con.Input.Button.Capture = 1
		case '	':
			con.Input.Button.ZL = 1
		case 'q':
			con.Input.Button.L = 1
		case ']':
			con.Input.Button.R = 1
		case '\\':
			con.Input.Button.ZL = 1
		case 'g':
			con.Input.Button.Plus = 1
		case 'f':
			con.Input.Button.Minus = 1
		default:
			log.Printf("unknown: %c = 0x%02x\n", buf[0], buf[0])
		}
		time.Sleep(50 * time.Millisecond)
		con.Input.Dpad.Left = 0
		con.Input.Dpad.Right = 0
		con.Input.Dpad.Up = 0
		con.Input.Dpad.Down = 0
		con.Input.Button.A = 0
		con.Input.Button.B = 0
		con.Input.Button.X = 0
		con.Input.Button.Y = 0
		con.Input.Button.L = 0
		con.Input.Button.R = 0
		con.Input.Button.ZL = 0
		con.Input.Button.ZR = 0
		con.Input.Button.Plus = 0
		con.Input.Button.Minus = 0
		con.Input.Button.Home = 0
		con.Input.Button.Capture = 0
	}
}
