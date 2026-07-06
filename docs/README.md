# The Seeed TinyGo Starter Kit

**kit**: a set of articles or equipment needed for a specific purpose.

Welcome! If a box with a tiny board and a pile of colorful modules just landed on your
desk, you are in exactly the right place. This is the official guide to the Seeed
TinyGo Starter Kit: a hands-on tour of embedded programming using the
[Go programming language](https://go.dev) compiled with [TinyGo](https://tinygo.org),
running on real hardware you can hold in your hand.

No prior electronics experience required. If you can read a Go program, you can make
things beep, blink, glow and measure the world around you by the end of the day.

<!-- TODO: photo of the full kit contents laid out -->

## What's in the box

- **Seeed Studio XIAO ESP32-C3** — a thumbnail-sized board with a RISC-V
  microcontroller, WiFi, Bluetooth and a USB-C port. This is the computer.
- **Grove Shield for XIAO** — a small carrier board the XIAO clicks into, with 8
  identical 4-pin sockets. This is how we avoid soldering and wiring mistakes.
- **Grove modules** — sensors and actuators, each in its own little PCB with a
  single socket. One cable, one socket, no way to plug it in backwards.

The modules covered by this guide:

buzzer, touch sensor, light sensor, rotary angle sensor, sound sensor (microphone),
temperature sensor, piezo vibration sensor, RGB LED stick, OLED display and a
3-axis accelerometer.

## The learning path

Each chapter introduces exactly **one new concept** and one module. They are ordered
so that every chapter builds on the previous ones — from wiggling a single output pin
all the way to driving devices over the I2C bus. Do them in order the first time;
jump around freely afterwards.

**Start here → [Setup: install the tools and flash your first program](setup.md)**

| # | Module | What you'll learn |
|---|--------|-------------------|
| 1 | [Buzzer](devices/01-buzzer.md) | Digital outputs: making sound by toggling a pin |
| 2 | [Touch sensor](devices/02-touch.md) | Digital inputs and pull resistors |
| 3 | [Light sensor](devices/03-light-sensor.md) | Analog signals and the ADC |
| 4 | [Rotary angle sensor](devices/04-rotary.md) | Potentiometers: mapping voltage to position |
| 5 | [Sound sensor](devices/05-sound.md) | Sampling signals fast: bias and loudness |
| 6 | [Temperature sensor](devices/06-temperature.md) | Thermistors: turning resistance into °C |
| 7 | [Piezo vibration sensor](devices/07-piezo.md) | Event detection with thresholds |
| 8 | [RGB LED stick](devices/08-rgb-led.md) | Addressable LEDs and frame buffers |
| 9 | [OLED display](devices/09-oled-display.md) | The I2C bus and pixel graphics |
| 10 | [Accelerometer](devices/10-accelerometer.md) | I2C sensors, registers and gravity |

After the tour, put it all together in the [projects](projects/README.md):
a house full of interacting devices, and a working audio recorder.

## How this guide works

Every chapter embeds a complete, tested program from this repository's
[`examples/`](../examples/) directory — the code you read on the page is the code
you flash, verbatim. Each chapter ends with **experiments**: small modifications to
try yourself. Do them. Until you change the code, do not expect to understand it.

## Takeaways

You have a computer that costs less than lunch, a compiler that speaks Go, and a
bag of sensors. The rest of this guide is just connecting those three things —
[let's set up](setup.md).
