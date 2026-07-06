# seeed-grove

Go support for the Seeed Grove ecosystem on XIAO boards, compiled with
[TinyGo](https://tinygo.org).

## 📦 Got the TinyGo Starter Kit? [**Start here →**](docs/README.md)

The [guide](docs/README.md) walks you from zero — no electronics experience —
through every module in the kit, one concept at a time, ending with full
multi-device projects.

## What's in this repository

- **[`grove`](grove.go)** (root package) — the Grove shield abstraction
  (`ShieldXiao`, `Conn`) and simple analog sensor types (light, rotary angle,
  sound, temperature, RGB strip). Works on the whole XIAO family:
  ESP32-C3/S3/C6, RP2040, RP2350, nRF52840 (ble).
- **[`drivers/`](drivers/)** — device drivers: [`lis3dh`](drivers/lis3dh/)
  3-axis accelerometer, [`ssd1306`](drivers/ssd1306/) OLED display with
  drawing and font support.
- **[`examples/`](examples/)** — runnable programs:
  [one per Grove module](examples/examples-grove-single/) plus
  [integration](examples/grove-clown-house/) and
  [audio streaming](examples/grove-sound-wav/) projects.
- **[`docs/`](docs/)** — the starter kit guide. Code blocks in the guide are
  spliced from `examples/` by [`internal/docgen`](internal/docgen/main.go);
  refresh with `go generate ./...`.

## Quick start

```sh
tinygo flash -target=xiao-esp32c3 ./examples/hello-world/
tinygo monitor
```

Using a different XIAO board? Swap the target: `xiao-esp32s3`, `xiao-rp2040`,
`xiao-rp2350`, `xiao-ble`.
