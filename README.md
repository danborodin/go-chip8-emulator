# CHIP-8 Emulator 
A CHIP-8 Emulator in golang


## Dependency

* golang 1.11.x and later (using go module)
* SDL2
  * https://github.com/veandco/go-sdl2

## Usage

* Run
  ```
  go run main.go -f /path/to/rom
  ```

* Reload game: press 'l' or 'L'


## Key Mapping
In this Emulator, CHIP-8 keys are mapped to below.
```
1 2 3 C ----> 1 2 3 4
4 5 6 D ----> Q W E R
7 8 9 E ----> A S D F
A 0 B F ----> Z X C V
```

## ROMs
ROMs are in the [games](./games) directory. 

## Reference
* https://en.wikipedia.org/wiki/CHIP-8
* http://mattmik.com/files/chip8/mastering/chip8.html
