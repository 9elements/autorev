# Protocol

## Shell

The DUT must have a simple shell driver, that is able to accept the following
commands:

```
enum {
	BL_START = 0,
	BL_FAKE,
	BL_FILTER,
	BL_CONFIG,
};
```

With the following types:
```
enum bl_io_type {
        MEM32,
        IO,
        MSR,
        CPUID,
	PCI,
};
```

encoded as
```
	MEM32 == m
	IO    == i
	MSR   == s
	CPUID == c
	PCI   == p

```

* Commands are send as ASCII hex dump.
* The shell identifier is `#B>`
* Commands are executed on '\n'
* Command ID and payload is separated by `: `

### BL_START

This command exists the shell and runs the following stage inside the blobolator.

Example:

`0: `

### BL_FAKE

**TODO: implement support for this**

Accepts the following payload:

```
struct {
	uint32_t ip;
	uint32_t address;
	uint64_t value;
	uint8_t inout;
	enum bl_io_type type;
};
```

The blobolator will match on `ip` and `address` and replace the read operation with
value `value`, if `inout` is 1.

The case for `inout` is 0 isn't specified yet.

### BL_FILTER

**TODO: implement support for this**

Accepts the following payload:

```
struct {
	uint32_t address;
	enum bl_io_type type;
};
```

### BL_CONFIG

Accepts the following payload:

```
struct {
	uint32_t length;
	uint8_t data[0]; // variable size
}
```

It's vendor dependend what should happen on this command.
It allows to change the DUT's runtime firmware configuration, for example BIOS POST options.

## Trace output

If not filtered the blobolator should output the trace in machine parseable format.

Every line is prefixed with

`#B!`

and has the following pattern:

`IP TYPE INOUT ADDR VALUE <VALUE2>`

where `IP`, `ADDR`, `VALUE`, `VALUE2` are 8 digits, TYPE is `m`, `i`, `s`,`c`,`p` INOUT is `I` or `O`.

Example:

```
#B! 07fe0001 m I ffffe000 00000000
```

* In case of `c` the ADDR is eax.
* In case of `P` the ADDR is the offset in MMCONF.
* In case of `s` VALUE2 is present, on all other instructions it's not.
* In case of `s` VALUE is EDX and VALUE2 is EAX, while ADDR is ECX.
