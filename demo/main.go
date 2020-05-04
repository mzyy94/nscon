// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"github.com/mzyy94/nscon"
	"log"
	"os"
	"os/exec"
	"time"
	"os/signal"
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
			con.Button.Dpad.Left = 1
		case 'd':
			con.Button.Dpad.Right = 1
		case 'w':
			con.Button.Dpad.Up = 1
		case 's':
			con.Button.Dpad.Down = 1
		case ' ':
			con.Button.Button.B = 1
		case 0x0a: // Enter
			con.Button.Button.A = 1
		case '.':
			con.Button.Button.X = 1
		case '/': 
			con.Button.Button.Y = 1
		case 0x1b: // Escape
			con.Button.Button.Home = 1
		case '`': 
			con.Button.Button.Capture = 1
		case '	':
			con.Button.Button.ZL = 1
		case 'q':
			con.Button.Button.L = 1
		case ']':
			con.Button.Button.R = 1
		case '\\':
			con.Button.Button.ZL = 1
		case 'g':
			con.Button.Button.Plus = 1
		case 'f':
			con.Button.Button.Minus = 1
		default:
			log.Printf("unknown: %c = 0x%02x\n", buf[0], buf[0])
		}
		time.Sleep(50 * time.Millisecond)
		con.Button.Dpad.Left = 0
		con.Button.Dpad.Right = 0
		con.Button.Dpad.Up = 0
		con.Button.Dpad.Down = 0
		con.Button.Button.A = 0
		con.Button.Button.B = 0
		con.Button.Button.X = 0
		con.Button.Button.Y = 0
		con.Button.Button.L = 0
		con.Button.Button.R = 0
		con.Button.Button.ZL = 0
		con.Button.Button.ZR = 0
		con.Button.Button.Plus = 0
		con.Button.Button.Minus = 0
		con.Button.Button.Home = 0
		con.Button.Button.Capture = 0
	}
}
