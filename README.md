# nscon

Nintendo Switch Controller simulator with usb gadget

## Features

_**Checked ones are currently supported.**_

- [x] ABXY Buttons Input
- [x] L/R/ZL/ZR Triggers Input
- [x] D-Pad Input
- [x] Meta Buttons Input
- [x] Left/Right Stick Input
- [x] Reconnection
- [ ] 6-Axis Accelerometer/Gyroscope
- [ ] Rumble Feedback
- [ ] LED Indicator
- [ ] Disconnection
- [ ] Remote Wakeup

## Usage

Create Nintendo Switch Pro Controller USB Gadget first.

e.g. https://gist.github.com/mzyy94/60ae253a45e2759451789a117c59acf9#file-add_procon_gadget-sh

### Simulate tty input as button input

```sh
sudo go run demo/main.go
```

## License

GPL 3.0 see [LICENSE](LICENSE)