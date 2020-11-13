package emulator

import "log"

const (
	Chip8DisplayW          = 64
	Chip8DisplayH          = 32
	CharacterSpritesOffset = 0x100
	CharacterSpriteBytes   = 5
	ProgrammOffset         = 0x200
	Chip8Frequency         = 60 * 8
)

type Chip8 struct {
	memory      [0x1000]uint8 //4096-byte RAM memory
	stack       [0x10]uint16  //Stack
	v           [0x10]uint8   //Registers
	keys        [0x10]uint8   //Key press state (HEX key code = index, 1 = pressed)
	dsp         [0x800]uint8  //Display (64 x 32 monochrome pixels), 0 or 1
	i           uint16        //Index register
	pc          uint16        //Program counter
	sp          uint8         //Stack pointer
	delayTimer  uint8         //Delay timer register
	soundTtimer uint8         //Sound timer regiter
}

var characterSprites = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

// Return pointer to a new Chip8 struct
func newChip8(b []byte) *Chip8 {
	c := &Chip8{}
	c.pc = ProgrammOffset
	c.sp = 0x0f

	//loading data to chip8
	copy(c.memory[ProgrammOffset:], []uint8(b))
	copy(c.memory[CharacterSpritesOffset:], characterSprites)
	return c
}

func (c *Chip8) step() {
	op := c.feetch()
	c.execute(op)
}

func (c *Chip8) decrementTimer() {
	if c.delayTimer > 0 {
		c.delayTimer--
	}
	if c.soundTtimer > 0 {
		c.soundTtimer--
	}
}

func (c *Chip8) feetch() uint16 {
	op := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	c.pc += 2
	return op
}

func (c *Chip8) updateCarryFlag(b bool) {
	if b {
		c.v[0xf] = 1
	} else {
		c.v[0xf] = 0
	}
}

func (c *Chip8) pushStak(v uint16) {
	c.stack[c.sp] = v
	c.sp--
}

func (c *Chip8) popStack() uint16 {
	c.sp++
	return c.stack[c.sp]
}

func (c *Chip8) pressedAnyKey() uint8 {
	for i, v := range c.keys {
		if v == 1 {
			return uint8(i)
		}
	}
	return 0xff
}

func (c *Chip8) execute(op uint16) {
	hex := op & 0xF000
	nnn := op & 0x0FFF
	nn := uint8(nnn & 0xFF)
	x := uint8((nnn >> 8) & 0xF)
	y := uint8((nnn >> 4) & 0xF)
	n := nn & 0x0f

	switch hexa {
	case 0x000:
		switch op {
		case 0x00E0:
			c.clearDisplay()
		case 0x00EE:
			c.returnFromSubroutine()
		default:
			log.Fatal("Not implemented 0NNN")
		}
	case 0x1000:
		c.jumpTo(nnn)
	case 0x2000:
		c.callSubroutine(nnn)
	case 0x3000:
		c.seByte(x, nn)
	case 0x4000:
		c.sneByte(x, nn)
	case 0x5000:
		c.seRegister(x, y)
	case 0x6000:
		c.loadByte(x, nn)
	case 0x7000:
		c.addByte(x, nn)
	case 0x8000:
		switch nnn & 0xf {
		case 0:
			c.loadReg(x, y)
		case 1:
			c.orReg(x, y)
		case 2:
			c.andReg(x, y)
		case 3:
			c.xorReg(x, y)
		case 4:
			c.addReg(x, y)
		case 5:
			c.subReg(x, y)
		case 6:
			c.shiftRight(x)
		case 7:
			c.subnReg(x, y)
		case 0xE:
			c.shiftLeft(x)
		}
	case 0x9000:
		c.sneReg(x, y)
	case 0xA000:
		c.loadI(nnn)
	case 0xB000:
		c.jumpV0(nnn)
	case 0xC000:
		c.random(x, nn)
	case 0xD000:
		c.draw(x, y, n)
	case 0xE000:
		switch nn {
		case 0x9E:
			c.skipKp(x)
		case 0xA1:
			c.skipNkp(x)
		}
	case 0xF000:
		switch nn {
		case 0x07:
			c.loadRegDelay(x)
		case 0x0A:
			c.loadRegKey(x)
		case 0x15:
			c.loadDt(x)
		case 0x18:
			c.loadSt(x)
		case 0x1E:
			c.addI(x)
		case 0x29:
			c.loadSprite(x)
		case 0x33:
			c.loadBCD(x)
		case 0x55:
			c.loadRegMem(x)
		case 0x65:
			c.loadMemReg(x)
		}
	}
}
