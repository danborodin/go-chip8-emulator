package main

import (
	"flag"
	"go-chip8-emulator/emulator"
	"io/ioutil"
	"os"
	"runtime"
)

var filename = flag.String("f", "", "chip8 image file path")

func init() {
	runtime.LockOSThread()
}

func main() {
	flag.Parse()

	f, _ := os.Open(*filename)
	binary, _ := ioutil.ReadAll(f)

	emu := emulator.NewEmulator(binary)
	emu.Run()
}
