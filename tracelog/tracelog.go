package tracelog

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial.v1"

	"github.com/9elements/autorev/config"
)

// LineType - The possible values of the ASCII string received from DUT
type LineType int

const (
	// MEM32 - 32bit memory access
	MEM32 LineType = iota
	// IO - IO port access
	IO
	// MSR - machine specific access
	MSR
	// CPUID - Read CPUID instruction
	CPUID
	// PCI - PCI register access
	PCI
)

// TraceLogEntry - Parse line into Line struct
type TraceLogEntry struct {
	// Instruction pointer
	IP uint
	// Use one of LineType
	Type int
	// True: In, False: Out
	Inout bool
	// The address that was accessed
	Address uint
	// The value read/written
	Value uint64
	// 8, 16, 32, 64 bit
	AccessSize uint
}

// TraceLog - Structure were we hold the general TraceLog Informations
type TraceLog struct {
	cfg config.Config
	// Database connection
	inputChannel chan string
	// serial connection
	serialConn serial.Port
	// fifo connection
	fifoFile1, fifoFile2 *os.File
	// verbosity
	verbose bool
}

// ConvertToType - Convert a string into a Type
func ConvertToType(inputType string) int {
	switch inputType {
	case "m":
		return int(MEM32)
	case "i":
		return int(IO)
	case "s":
		return int(MSR)
	case "c":
		return int(CPUID)
	case "p":
		return int(PCI)
	}
	return -1
}

// ConvertToDir - Convert InputDirection into Bool
func ConvertToDir(inputDirection string) bool {
	switch inputDirection {
	case "I":
		return true
	case "O":
		return false
	}
	return false
}

// String - Convert TraceLogEntry into a Readable String
func (tle *TraceLogEntry) String() string {
	var dir string
	if tle.Inout {
		dir = "in"
	} else {
		dir = "out"
	}

	var tracelogtype string
	switch tle.Type {
	case int(MEM32):
		tracelogtype = "m"
	case int(IO):
		tracelogtype = "i"
	case int(MSR):
		tracelogtype = "s"
	case int(CPUID):
		tracelogtype = "c"
	case int(PCI):
		tracelogtype = "p"
	}
	return fmt.Sprintf("IP: %08x, Type: %s, Dir: %s, Addr: %08x, Value: %016x, Access: %d",
		tle.IP, tracelogtype, dir, tle.Address, tle.Value, tle.AccessSize)
}

// SmallString - Convert TraceLogEntry into a Readable String
func (tle *TraceLogEntry) SmallString() string {
	var dir string
	if tle.Inout {
		dir = "<-"
	} else {
		dir = "->"
	}

	var tracelogtype string
	switch tle.Type {
	case int(MEM32):
		tracelogtype = "m"
	case int(IO):
		tracelogtype = "i"
	case int(MSR):
		tracelogtype = "s"
	case int(CPUID):
		tracelogtype = "c"
	case int(PCI):
		tracelogtype = "p"
	}
	return fmt.Sprintf("%s%s %08x %016x %d",
		tracelogtype, dir, tle.Address, tle.Value, tle.AccessSize)
}

// ParseLine - Parse a new Line
func ParseLine(line string) (*TraceLogEntry, error) {
	var new TraceLogEntry
	parts := strings.Split(line, " ")

	if len(parts) == 0 {
		return nil, fmt.Errorf("Empty line")
	}
	if parts[0] != "#B!" {
		return nil, fmt.Errorf("Line doesn't start with trace prefix")
	}
	if len(parts) == 1 {
		return nil, fmt.Errorf("Line doesn't have IP value")
	}
	i, err := strconv.ParseInt(parts[1], 16, 64)
	if err != nil {
		return nil, err
	}
	new.IP = uint(i)

	if len(parts) == 2 {
		return nil, fmt.Errorf("Line doesn't have Type value")
	}
	t := ConvertToType(parts[2])
	if t == -1 {
		return nil, fmt.Errorf("Line has unknown type value")
	}
	new.Type = t

	if len(parts) == 3 {
		return nil, fmt.Errorf("Line doesn't have InOut value")
	}
	if parts[3] != "I" && parts[3] != "O" {
		return nil, fmt.Errorf("Line has unknown InOut value")
	}
	new.Inout = parts[3] == "I"

	if len(parts) == 4 {
		return nil, fmt.Errorf("Line doesn't have Address value")
	}
	i, err = strconv.ParseInt(parts[4], 16, 64)
	if err != nil {
		return nil, err
	}
	new.Address = uint(i)

	if len(parts) == 5 {
		return nil, fmt.Errorf("Line doesn't have Value value")
	}
	i, err = strconv.ParseInt(parts[5], 16, 64)
	if err != nil {
		return nil, err
	}
	new.Value = uint64(i)

	if t == int(MSR) {
		if len(parts) == 6 {
			return nil, fmt.Errorf("Line doesn't have second Value value")
		}
		i, err = strconv.ParseInt(parts[6], 16, 64)
		if err != nil {
			return nil, err
		}

		new.Value |= uint64(i) << 32
	}

	if t == int(IO) || t == int(MEM32) || t == int(PCI) {
		if len(parts) == 6 {
			return nil, fmt.Errorf("Line doesn't have AccessSize value")
		}
		i, err = strconv.ParseInt(parts[6], 10, 32)
		if err != nil {
			return nil, err
		}
		new.AccessSize = uint(i)
	}

	return &new, nil
}

// CreateTraceLog - Create a new Trace Log
func CreateTraceLog(devTTYDevicePath string, baudTTYDevice int, fifoDevicePath string, cfg config.Config) (*TraceLog, error) {
	var t TraceLog

	// argument error handling
	if len(devTTYDevicePath) > 0 && len(fifoDevicePath) > 0 {
		return nil, fmt.Errorf("Cannot specify both: -fifo and -dev")
	}
	if len(cfg.TraceLog.Serial.Port) == 0 {
		return nil, fmt.Errorf("Collecting a new trace, but serial port not specified")
	}
	if cfg.TraceLog.Serial.Type != "fifo" && cfg.TraceLog.Serial.Type != "tty" {
		return nil, fmt.Errorf("Collecting a new trace, but serial type is unknown (must be 'fifo' or 'tty')")
	}

	t.cfg = cfg

	// Execute shell command to get serial and DUT control
	if len(t.cfg.TraceLog.DutControl.InitCmd) > 0 {
		t.shellcmd(t.cfg.TraceLog.DutControl.InitCmd)
	}

	// overwrite config
	if len(devTTYDevicePath) > 0 {
		t.cfg.TraceLog.Serial.Type = "tty"
		t.cfg.TraceLog.Serial.Port = devTTYDevicePath
	}
	if len(fifoDevicePath) > 0 {
		t.cfg.TraceLog.Serial.Type = "fifo"
		t.cfg.TraceLog.Serial.Port = fifoDevicePath
	}
	if baudTTYDevice > 0 {
		t.cfg.TraceLog.Serial.BaudRate = baudTTYDevice
	}

	return &t, nil
}

// SetVerbose - Set verbosity of tracing
func (tl *TraceLog) SetVerbose(v bool) {
	tl.verbose = v
}

// shellcmd - execute a shell cmd. Allows to run custom scripts that power the DUT
func (tl *TraceLog) shellcmd(s string) {
	cmd := strings.Split(s, " ")[0]
	args := strings.Split(s, " ")[1:]
	abs, err := filepath.Abs(cmd)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Running shell cmd '%s'\n", abs)

		execcmd := exec.Command(abs, args...)
		execcmd.Dir = filepath.Dir(abs)

		err := execcmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// openWaitForSerial - Wait for the serial device to appear
func (tl *TraceLog) openWaitForSerial(timeout uint) error {
	var err error
	n := time.Now()
	if timeout == 0 {
		timeout = 5
	}
	limit := time.Duration(timeout) * time.Second

	if tl.cfg.TraceLog.Serial.ReadWriteTimeout == 0 {
		tl.cfg.TraceLog.Serial.ReadWriteTimeout = 5
	}

	if tl.cfg.TraceLog.Serial.Type == "tty" {
		if tl.cfg.TraceLog.Serial.BaudRate == 0 {
			tl.cfg.TraceLog.Serial.BaudRate = 115200
		}
		var mode = serial.Mode{
			tl.cfg.TraceLog.Serial.BaudRate, 8, serial.NoParity, serial.OneStopBit,
		}
		for time.Since(n) < limit {
			// Poll on Serial to open (Testing)
			tl.serialConn, err = serial.Open(tl.cfg.TraceLog.Serial.Port, &mode)
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond)
		}
		if time.Since(n) >= limit {
			tl.serialConn = nil
			return fmt.Errorf("Timeout waiting for serial device")
		}
	} else if tl.cfg.TraceLog.Serial.Type == "fifo" {
		for time.Since(n) < limit {
			// Poll on FIFO to open (Testing)
			var res error
			c1 := make(chan error, 1)
			go func() {
				tl.fifoFile1, err = os.OpenFile(tl.cfg.TraceLog.Serial.Port+".in", os.O_WRONLY, 0)
				if err == nil {
					tl.fifoFile2, err = os.OpenFile(tl.cfg.TraceLog.Serial.Port+".out", os.O_RDONLY, 0)
					if err != nil {
						tl.fifoFile1.Close()
					}
				}
				c1 <- err
			}()
			log.Printf("Waiting for open file\n")

			select {
			case res = <-c1:
				{
					if res != nil {
						return res
					}
					if tl.fifoFile2 == nil || tl.fifoFile1 == nil {
						return fmt.Errorf("Failed to open FIFO")
					}
					return nil
				}
			case <-time.After(time.Second):
				{
				}
			}
		}
		if time.Since(n) >= limit {
			return fmt.Errorf("Timeout waiting for fifo device")
		}
	}

	return nil
}

// read - Reads single char from serial or FIFO with timeout
func (tl *TraceLog) read() (byte, error) {
	var res byte
	limit := time.Duration(tl.cfg.TraceLog.Serial.ReadWriteTimeout) * time.Second

	c1 := make(chan byte, 1)
	go func() {
		b := make([]byte, 1)
		if tl.cfg.TraceLog.Serial.Type == "tty" {
			n, err := tl.serialConn.Read(b)
			if err == nil && n == 1 {
				c1 <- b[0]
			}
		} else {
			n, err := tl.fifoFile2.Read(b)
			if err == nil && n > 0 {
				c1 <- b[0]
			}
		}
	}()

	select {
	case res = <-c1:
		{
			return res, nil
		}
	case <-time.After(limit):
		{
		}
	}

	return 0, fmt.Errorf("Timeout waiting for serial char")
}

// readLine - Reads line from serial or FIFO with timeout
// Cuts of newline "\n"
func (tl *TraceLog) readLine() (string, error) {
	var res string
	limit := time.Duration(tl.cfg.TraceLog.Serial.ReadWriteTimeout) * time.Second

	n := time.Now()

	for time.Since(n) < limit {
		b, err := tl.read()
		if err != nil {
			return "", err
		}
		if b == '\r' {
			b, err = tl.read()
			if err != nil {
				return "", err
			}
		}
		if b == '\n' {
			return res, nil
		}
		res += string(b)
	}

	return "", fmt.Errorf("Timeout waiting for line from serial device")
}

// read from serial until search string search is found
func (tl *TraceLog) readString(search string) (string, error) {
	var res string
	limit := time.Duration(tl.cfg.TraceLog.Serial.ReadWriteTimeout) * time.Second

	n := time.Now()

	for time.Since(n) < limit {
		b, err := tl.read()
		if err != nil {
			return "", fmt.Errorf("Timeout waiting for full serial line")
		}
		res += string(b)
		if strings.Contains(res, search) {
			return res, nil
		}
	}

	return "", fmt.Errorf("Timeout waiting for serial device")
}

// Write to Serial or FIFO with timeout
func (tl *TraceLog) write(buf []byte) (int, error) {
	limit := time.Duration(tl.cfg.TraceLog.Serial.ReadWriteTimeout) * time.Second

	n := time.Now()

	c := 0

	for time.Since(n) < limit {

		c1 := make(chan bool, 1)
		go func(b byte) {
			if tl.cfg.TraceLog.Serial.Type == "tty" {
				n, err := tl.serialConn.Write([]byte{b})
				if err == nil && n == 1 {
					c1 <- true
				}
			} else {
				n, err := tl.fifoFile1.Write([]byte{b})
				if err == nil && n == 1 {
					c1 <- true
				}
			}
		}(buf[0])

		select {
		case <-c1:
			{
				buf = buf[1:]
				c++
			}
		case <-time.After(limit):
			{
				return c, fmt.Errorf("Timeout waiting for serial char")
			}
		}
		if len(buf) == 0 {
			break
		}
	}
	if time.Since(n) >= limit {
		return c, fmt.Errorf("Timeout waiting to send serial char")
	}

	return c, nil
}

// See write
func (tl *TraceLog) writeString(s string) (int, error) {
	return tl.write([]byte(s))
}

// Write to Serial or FIFO with timeout
func (tl *TraceLog) close() {
	if tl.cfg.TraceLog.Serial.Type == "tty" {
		tl.serialConn.Close()
		tl.serialConn = nil
	} else {
		if tl.fifoFile1 != nil {
			tl.fifoFile1.Close()
			tl.fifoFile1 = nil

		}
		if tl.fifoFile2 != nil {
			tl.fifoFile2.Close()
			tl.fifoFile2 = nil
		}
	}
}

// CollectNewTracelog - Collects a new Trace Log
func (tl *TraceLog) CollectNewTracelog(config []byte) ([]TraceLogEntry, error) {
	var ret []TraceLogEntry
	var err error

	// Execute shell command to get DUT in the running state
	if len(tl.cfg.TraceLog.DutControl.StartCmd) > 0 {
		tl.shellcmd(tl.cfg.TraceLog.DutControl.StartCmd)
	}
	defer func() {
		// Execute shell command to get DUT in the off state
		if len(tl.cfg.TraceLog.DutControl.StopCmd) > 0 {
			tl.shellcmd(tl.cfg.TraceLog.DutControl.StopCmd)
		}
	}()

	// Need to open serial and FIFO here as
	// 1) xhci debug serial appears only when the DUT enabled the debug port
	// 2) the qemu FIFO blocks until the other end of the FIFO is opened

	err = tl.openWaitForSerial(tl.cfg.TraceLog.Serial.DeviceHotplugTimeout)
	if err != nil {
		return nil, err
	}
	defer tl.close()

	// Wait for shell to connect
	for {
		tl.write([]byte("\n"))
		// read shell prefix
		l, err := tl.readString(">")
		if err != nil {
			return ret, fmt.Errorf("Failed to parse shell prefix")
		}
		if tl.verbose {
			s := strings.Split(l, "\n")
			for i := range s {
				log.Printf(">%s<\n", s[i])
			}
		}
		if strings.HasSuffix(l, "#B>") {
			break
		}
	}
	log.Printf("Found shell...\n")
	// Set config
	offset := 0
	remain := len(config)
	for offset < 4096 && offset < remain { // hardcoded in libx86emu
		tl.writeString("3: ")
		if remain > 512 {
			tl.writeString(fmt.Sprintf("%08x", 512))
			tl.writeString(fmt.Sprintf("%08x", offset))
			for i := 0; i < 512; i++ {
				tl.writeString(fmt.Sprintf("%02x", config[i+offset]))
			}

			offset += 512
			remain -= 512

		} else {
			tl.writeString(fmt.Sprintf("%08x", remain))
			tl.writeString(fmt.Sprintf("%08x", offset))
			for i := 0; i < remain; i++ {
				tl.writeString(fmt.Sprintf("%02x", config[i+offset]))
			}

			offset += remain
			remain = 0
		}
		tl.writeString("\n")

		discard, err := tl.readString(">")
		if err != nil {
			return nil, err
		}
		if tl.verbose {
			s := strings.Split(discard, "\n")
			for i := range s {
				log.Printf(">%s<\n", s[i])
			}
		}
	}

	// Write start signal
	tl.writeString("0: \n")

	log.Println("Config written.")

	checkCaptureState := len(tl.cfg.TraceLog.StartSignal.Type) > 0 && len(tl.cfg.TraceLog.StopSignal.Type) > 0
	capturing := !checkCaptureState

	var running = true
	for running {
		buffer, err := tl.readLine()
		if err != nil {
			return ret, fmt.Errorf("Failed to read from DUT connection: %s", err.Error())
		}

		inputLog, err := ParseLine(buffer)
		if err != nil {
			if tl.verbose {
				log.Printf(">%s<\n", buffer)
			}
			if capturing {
				log.Printf("Error ! %v\n", err)
				log.Printf("Line was >>%s<<\n", buffer)
			}
		} else {
			if checkCaptureState {
				if capturing {
					ret = append(ret, *inputLog)
					log.Printf("%v\n", inputLog)
					if inputLog.Inout == ConvertToDir(tl.cfg.TraceLog.StopSignal.Direction) &&
						inputLog.AccessSize == tl.cfg.TraceLog.StopSignal.DataWidth &&
						inputLog.Value == tl.cfg.TraceLog.StopSignal.Value &&
						inputLog.Type == ConvertToType(tl.cfg.TraceLog.StopSignal.Type) &&
						inputLog.Address == tl.cfg.TraceLog.StopSignal.Offset {
						capturing = false
						running = false
						break
					}
				}
				if !capturing {
					if inputLog.Inout == ConvertToDir(tl.cfg.TraceLog.StartSignal.Direction) &&
						inputLog.AccessSize == tl.cfg.TraceLog.StartSignal.DataWidth &&
						inputLog.Value == tl.cfg.TraceLog.StartSignal.Value &&
						inputLog.Type == ConvertToType(tl.cfg.TraceLog.StartSignal.Type) &&
						inputLog.Address == tl.cfg.TraceLog.StartSignal.Offset {
						capturing = true
						log.Printf("%v\n", inputLog)
					}
				}
			} else {
				ret = append(ret, *inputLog)
				log.Printf("%v\n", inputLog)
			}
		}
	}
	return ret, nil
}
