package emulator

import (
	"math/rand"
	"time"
)

// Seed rand function.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Clears the screen.
func (c *Chip8) clearDisplay() {
	for i := range c.dsp {
		c.dsp[i] = 0
	}
}

// Returns from a subroutine.
func (c *Chip8) returnFromSubroutine() {
	r := c.popStack()
	c.pc = r
}

// Jumps to address NNN.
func (c *Chip8) jumpTo(nnn uint16) {
	c.pc = nnn
}

// Calls subroutine at NNN.
func (c *Chip8) callSubroutine(nnn uint16) {
	c.pushStak(c.pc)
	c.pc = nnn
}

// Skips the next instruction if VX equals NN.
// (Usually the next instruction is a jump to skip a code block)
func (c *Chip8) seByte(x, nn uint8) {
	if c.v[x] == nn {
		c.pc += 2
	}
}

// Skips the next instruction if VX doesn't equal NN.
// (Usually the next instruction is a jump to skip a code block)
func (c *Chip8) sneByte(x, nn uint8) {
	if c.v[x] != nn {
		c.pc += 2
	}
}

// Skips the next instruction if VX equals VY.
// (Usually the next instruction is a jump to skip a code block)
func (c *Chip8) seRegister(x, y uint8) {
	if c.v[x] == c.v[y] {
		c.pc += 2
	}
}

// Sets VX to NN.
func (c *Chip8) loadByte(x, nn uint8) {
	c.v[x] = nn
}

// Adds NN to VX. (Carry flag is not changed)
func (c *Chip8) addByte(x, nn uint8) {
	c.v[x] += nn
}

// Sets VX to the value of VY.
func (c *Chip8) loadReg(x, y uint8) {
	c.v[x] = c.v[y]
}

// Sets VX to VX or VY. (Bitwise OR operation)
func (c *Chip8) orReg(x, y uint8) {
	c.v[x] |= c.v[y]
}

// Sets VX to VX and VY. (Bitwise AND operation)
func (c *Chip8) andReg(x, y uint8) {
	c.v[x] &= c.v[y]
}

// Sets VX to VX xor VY.
func (c *Chip8) xorReg(x, y uint8) {
	c.v[x] ^= c.v[y]
}

// Adds VY to VX. VF is set to 1 when there's a carry, and to 0
// when there isn't.
func (c *Chip8) addReg(x, y uint8) {
	carried := (uint16(c.v[x]) + uint16(c.v[y])) > 0xff
	c.v[x] += c.v[y]
	c.updateCarryFlag(carried)
}

// VY is subtracted from VX. VF is set to 0 when there's a borrow,
// and 1 when there isn't.
func (c *Chip8) subReg(x, y uint8) {
	borrowed := c.v[x] < c.v[y]
	c.v[x] -= c.v[y]
	c.updateCarryFlag(!borrowed)
}

// Stores the least significant bit of VX in VF and then shifts
// VX to the right by 1.
func (c *Chip8) shiftRight(x uint8) {
	c.updateCarryFlag((c.v[x] & 0x01) == 1)
	c.v[x] = c.v[x] >> 1
}

// Sets VX to VY minus VX. VF is set to 0 when there's a borrow,
// and 1 when there isn't.
func (c *Chip8) subnReg(x, y uint8) {
	borrowed := c.v[y] < c.v[x]
	c.v[x] = c.v[y] - c.v[x]
	c.updateCarryFlag(!borrowed)
}

// Stores the most significant bit of VX in VF and then shifts
// VX to the left by 1.
func (c *Chip8) shiftLeft(x uint8) {
	c.updateCarryFlag((c.v[x] >> 7) == 1)
	c.v[x] = c.v[x] << 1
}

// Skips the next instruction if VX doesn't equal VY.
// (Usually the next instruction is a jump to skip a code block)
func (c *Chip8) sneReg(x, y uint8) {
	if c.v[x] != c.v[y] {
		c.pc += 2
	}
}

// Sets I to the address NNN.
func (c *Chip8) loadI(nnn uint16) {
	c.i = nnn
}

// Jumps to the address NNN plus V0.
func (c *Chip8) jumpV0(nnn uint16) {
	c.pc = uint16(c.v[0]) + nnn
}

// Sets VX to the result of a bitwise and operation on a random
// number (Typically: 0 to 255) and NN.
func (c *Chip8) random(x, nn uint8) {
	c.v[x] = uint8(rand.Uint32() & uint32(nn))
}

// Draws a sprite at coordinate (VX, VY)
// that has a width of 8 pixels and a height of N+1 pixels.
// Each row of 8 pixels is read as bit-coded starting from
// memory location I; I value doesn’t change after the execution
// of this instruction. As described above, VF is set to 1 if any
// screen pixels are flipped from set to unset when the sprite is
// drawn, and to 0 if that doesn’t happen
func (c *Chip8) draw(x, y, n uint8) {
	x = c.v[x]
	y = c.v[y]
	flipped := false
	sm := c.memory[c.i:] // sprite address memory
	for iy := uint8(0); iy < n; iy++ {
		for ix := uint8(0); ix < 8; ix++ {
			tx := int(x) + int(ix)
			ty := int(y) + int(iy)
			if tx >= Chip8DisplayW || ty >= Chip8DisplayH {
				continue
			}

			s := c.dsp[ty*Chip8DisplayW+tx]
			d := (sm[iy] >> (7 - ix)) & 0x01
			c.dsp[ty*Chip8DisplayW+tx] ^= d
			if s == 1 && d == 1 {
				flipped = true
			}
		}
	}
	c.updateCarryFlag(flipped)
}

// Skips the next instruction if the key stored in VX is pressed.
// (Usually the next instruction is a jump to skip a code block)
func (c *Chip8) skipKp(x uint8) {
	if c.keys[c.v[x]] == 1 {
		c.pc += 2
	}
}

// Skips the next instruction if the key stored in VX isn't pressed.
// (Usually the next instruction is a jump to skip a code block)
func (c *Chip8) skipNkp(x uint8) {
	if c.keys[c.v[x]] == 0 {
		c.pc += 2
	}
}

// Sets VX to the value of the delay timer.
func (c *Chip8) loadRegDelay(x uint8) {
	c.v[x] = c.delayTimer
}

// A key press is awaited, and then stored in VX.
// (Blocking Operation. All instruction halted until next key event)
func (c *Chip8) loadRegKey(x uint8) {
	if c.pressedAnyKey() == 0xff {
		c.pc -= 2
	} else {
		c.v[x] = c.pressedAnyKey()
	}
}

// Sets the delay timer to VX.
func (c *Chip8) loadDt(x uint8) {
	c.delayTimer = c.v[x]
}

// Sets the sound timer to VX.
func (c *Chip8) loadSt(x uint8) {
	c.soundTtimer = c.v[x]
}

// Adds VX to I. VF is not affected.
func (c *Chip8) addI(x uint8) {
	c.i += uint16(c.v[x])
}

// Sets I to the location of the sprite for the character in VX.
// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
func (c *Chip8) loadSprite(x uint8) {
	c.i = CharacterSpritesOffset + uint16(c.v[x])*CharacterSpriteBytes
}

// Stores the binary-coded decimal representation of VX, with
// the most significant of three digits at the address in I,
// the middle digit at I plus 1, and the least significant digit
// at I plus 2. (In other words, take the decimal representation
// of VX, place the hundreds digit in memory at location in I,
// the tens digit at location I+1, and the ones digit at location
// I+2.)
func (c *Chip8) loadBCD(x uint8) {
	c.memory[c.i+0] = c.v[x] / 100
	c.memory[c.i+1] = (c.v[x] % 100) / 10
	c.memory[c.i+2] = c.v[x] % 10
}

// Stores V0 to VX (including VX) in memory starting at address I.
//  The offset from I is increased by 1 for each value written,
// but I itself is left unmodified.
func (c *Chip8) loadRegMem(x uint8) {
	copy(c.memory[c.i:], c.v[:x+1])
}

// Fills V0 to VX (including VX) with values from memory starting
// at address I. The offset from I is increased by 1 for each value
// written, but I itself is left unmodified.
func (c *Chip8) loadMemReg(x uint8) {
	copy(c.v[:x+1], c.memory[c.i:])
}
