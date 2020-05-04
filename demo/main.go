// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"log"
	"os"
	"os/exec"
	"time"
	"github.com/mzyy94/nscon"
)

func main() {
	target := "/dev/hidg0"
	con := nscon.NewController(target)
	defer con.Close()
	con.Connect()

	buf := make([]byte, 1)

	// Set tty break for read keyboard input directly
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	for {
		os.Stdin.Read(buf)
		switch buf[0] {
		case 0x61:
			con.Button.Dpad.Left = 1
		case 0x64:
			con.Button.Dpad.Right = 1
		case 0x77:
			con.Button.Dpad.Up = 1
		case 0x73:
			con.Button.Dpad.Down = 1
		default:
			log.Printf("unknown: %c = 0x%02x\n", buf[0], buf[0])
		}
		time.Sleep(50 * time.Millisecond)
		con.Button.Dpad.Left = 0
		con.Button.Dpad.Right = 0
		con.Button.Dpad.Up = 0
		con.Button.Dpad.Down = 0
	}
}
