// SPDX-License-Identifier: GPL-3.0-only

package nscon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

var SPI_ROM_DATA = map[byte][]byte{
	0x60: []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0x03, 0xa0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02, 0xff, 0xff, 0xff, 0xff,
		0xf0, 0xff, 0x89, 0x00, 0xf0, 0x01, 0x00, 0x40, 0x00, 0x40, 0x00, 0x40, 0xf9, 0xff, 0x06, 0x00,
		0x09, 0x00, 0xe7, 0x3b, 0xe7, 0x3b, 0xe7, 0x3b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xba, 0x15, 0x62,
		0x11, 0xb8, 0x7f, 0x29, 0x06, 0x5b, 0xff, 0xe7, 0x7e, 0x0e, 0x36, 0x56, 0x9e, 0x85, 0x60, 0xff,
		0x32, 0x32, 0x32, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x50, 0xfd, 0x00, 0x00, 0xc6, 0x0f, 0x0f, 0x30, 0x61, 0x96, 0x30, 0xf3, 0xd4, 0x14, 0x54, 0x41,
		0x15, 0x54, 0xc7, 0x79, 0x9c, 0x33, 0x36, 0x63, 0x0f, 0x30, 0x61, 0x96, 0x30, 0xf3, 0xd4, 0x14,
		0x54, 0x41, 0x15, 0x54, 0xc7, 0x79, 0x9c, 0x33, 0x36, 0x63, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	},
	0x80: []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xb2, 0xa1, 0xbe, 0xff, 0x3e, 0x00, 0xf0, 0x01, 0x00, 0x40,
		0x00, 0x40, 0x00, 0x40, 0xfe, 0xff, 0xfe, 0xff, 0x08, 0x00, 0xe7, 0x3b, 0xe7, 0x3b, 0xe7, 0x3b,
	},
}

type Gadget struct {
	name string
}

func (g Gadget) state() bool {
	buf, err := ioutil.ReadFile(fmt.Sprintf("/sys/kernel/config/usb_gadget/%s/UDC", g.name))
	if err != nil {
		return false
	}

	return len(buf) > 1
}

func (g Gadget) enable() error {
	udcs, err := ioutil.ReadDir("/sys/class/udc")
	if err != nil {
		return err
	}
	if len(udcs) == 0 {
		return errors.New("UDC not found")
	}
	return ioutil.WriteFile(fmt.Sprintf("/sys/kernel/config/usb_gadget/%s/UDC", g.name),
		[]byte(udcs[0].Name()), os.ModeCharDevice)
}

func (g Gadget) disable() error {
	return ioutil.WriteFile(fmt.Sprintf("/sys/kernel/config/usb_gadget/%s/UDC", g.name),
		[]byte{0x0a}, os.ModeCharDevice)
}

type ControllerInput struct {
	Dpad struct {
		Up, Down, Left, Right uint8
	}
	Button struct {
		A, B, X, Y, R, ZR, L, ZL   uint8
		Home, Plus, Minus, Capture uint8
	}
	Stick struct {
		Left, Right struct {
			X, Y  float64
			Press uint8
		}
	}
}

type Controller struct {
	path            string
	fp              *os.File
	gadget          Gadget
	count           uint8
	stopCounter     chan struct{}
	stopInput       chan struct{}
	stopCommunicate chan struct{}
	Input           ControllerInput
	LogLevel        int
}

// NewController create an instance of Controller with device path
func NewController(path string, name string) *Controller {
	gadget := Gadget{name}
	return &Controller{
		path:   path,
		gadget: gadget,
	}
}

// Close closes all channel and device file
func (c *Controller) Close() {
	if c.fp == nil {
		if c.LogLevel > 0 {
			log.Println("Already closed.")
		}
		return
	}
	close(c.stopCounter)
	close(c.stopInput)
	close(c.stopCommunicate)
	c.fp.Close()
	c.fp = nil
	c.gadget.disable()
}

func (c *Controller) startCounter() {
	ticker := time.NewTicker(time.Millisecond * 5)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.count++
			case <-c.stopCounter:
				return
			}
		}
	}()
}

func packShorts(short1, short2 uint16) (data []byte) {
	data = make([]byte, 3)
	data[0] = uint8(short1 & 0xff)
	data[1] = uint8(((short2 << 4) & 0xf0) | ((short1 >> 8) & 0x0f))
	data[2] = uint8((short2 >> 4) & 0xff)
	return data
}

func (c *Controller) getInputBuffer() []byte {
	left := c.Input.Button.Y |
		c.Input.Button.X<<1 |
		c.Input.Button.B<<2 |
		c.Input.Button.A<<3 |
		c.Input.Button.R<<6 |
		c.Input.Button.ZR<<7

	center := c.Input.Button.Minus |
		c.Input.Button.Plus<<1 |
		c.Input.Stick.Right.Press<<2 |
		c.Input.Stick.Left.Press<<3 |
		c.Input.Button.Home<<4 |
		c.Input.Button.Capture<<5

	right := c.Input.Dpad.Down |
		c.Input.Dpad.Up<<1 |
		c.Input.Dpad.Right<<2 |
		c.Input.Dpad.Left<<3 |
		c.Input.Button.L<<6 |
		c.Input.Button.ZL<<7

	lx := uint16(math.Round((1 + c.Input.Stick.Left.X) * 2048))
	ly := uint16(math.Round((1 + c.Input.Stick.Left.Y) * 2048))
	rx := uint16(math.Round((1 + c.Input.Stick.Right.X) * 2048))
	ry := uint16(math.Round((1 + c.Input.Stick.Right.Y) * 2048))

	leftStick := packShorts(lx, ly)
	rightStick := packShorts(rx, ry)

	return []byte{0x81, left, center, right, leftStick[0], leftStick[1],
		leftStick[2], rightStick[0], rightStick[1], rightStick[2], 0x00}
}

func (c *Controller) startInputReport() {
	ticker := time.NewTicker(time.Millisecond * 30)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.write(0x30, c.count, c.getInputBuffer())
			case <-c.stopInput:
				return
			}
		}
	}()
}

func (c *Controller) uart(ack bool, subCmd byte, data []byte) {
	ackByte := byte(0x00)
	if ack {
		ackByte = 0x80
		if len(data) > 0 {
			ackByte |= subCmd
		}
	}
	c.write(0x21, c.count, append(append(c.getInputBuffer(), []byte{ackByte, subCmd}...), data...))
}

func (c *Controller) write(ack byte, cmd byte, buf []byte) {
	data := append(append([]byte{ack, cmd}, buf...), make([]byte, 62-len(buf))...)
	c.fp.Write(data)
	if c.LogLevel > 0 && ack != 0x30 {
		log.Println("write:", hex.EncodeToString(data))
	}
}

// Connect begins connection to device
func (c *Controller) Connect() error {
	var err error
	if c.fp != nil {
		return errors.New("Already connected.")
	}

	if c.gadget.name != "" && c.gadget.state() == false {
		c.gadget.enable()
	}
	c.fp, err = os.OpenFile(c.path, os.O_RDWR|os.O_SYNC, os.ModeDevice)
	if err != nil {
		return err
	}

	c.stopCounter = make(chan struct{})
	c.stopInput = make(chan struct{})
	c.stopCommunicate = make(chan struct{})

	c.startCounter()
	go func() {
		buf := make([]byte, 128)

		for {
			select {
			case <-c.stopCommunicate:
				return
			default:
			}

			n, err := c.fp.Read(buf)
			if c.LogLevel > 0 {
				log.Println("read:", hex.EncodeToString(buf[:n]), err)
			}
			switch buf[0] {
			case 0x80:
				switch buf[1] {
				case 0x01:
					c.write(0x81, buf[1], []byte{0x00, 0x03, 0x00, 0x00, 0x5e, 0x00, 0x53, 0x5e})
				case 0x02, 0x03:
					c.write(0x81, buf[1], []byte{})
				case 0x04:
					c.startInputReport()
				case 0x05:
					close(c.stopInput)
					c.stopInput = make(chan struct{})
				}
			case 0x01:
				switch buf[10] {
				case 0x01: // Bluetooth manual pairing
					c.uart(true, buf[10], []byte{0x03, 0x01})
				case 0x02: // Request device info
					c.uart(true, buf[10], []byte{0x03, 0x48, 0x03,
						0x02, 0x5e, 0x53, 0x00, 0x5e, 0x00, 0x00, 0x03, 0x01})
				case 0x03, 0x08, 0x30, 0x38, 0x40, 0x41, 0x48: // Empty response
					c.uart(true, buf[10], []byte{})
				case 0x04: // Empty response
					c.uart(true, buf[10], []byte{})
				case 0x10: // Read SPI ROM
					data, ok := SPI_ROM_DATA[buf[12]]
					if ok {
						c.uart(true, buf[10], append(buf[11:16],
							data[buf[11]:buf[11]+buf[15]]...))
						if c.LogLevel > 1 {
							log.Printf("Read SPI address: %02x%02x[%d] %v\n",
								buf[12], buf[11], buf[15], data[buf[11]:buf[11]+buf[15]])
						}
					} else {
						c.uart(false, buf[10], []byte{})
						if c.LogLevel > 1 {
							log.Printf("Unknown SPI address: %02x[%d]\n", buf[12], buf[15])
						}
					}
				case 0x21:
					// FIXME: Check ack value
					c.uart(true, buf[10], []byte{0x01, 0x00, 0xff, 0x00, 0x03, 0x00, 0x05, 0x01})
				default:
					if c.LogLevel > 1 {
						log.Println("UART unknown request", buf[10], buf)
					}
				}

			case 0x00:
			case 0x10:
			default:
				if c.LogLevel > 1 {
					log.Println("unknown request", buf[0])
				}
			}
		}
	}()

	return nil
}
