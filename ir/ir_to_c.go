package ir

import "fmt"

func (p PRead) ConvertToC() string {
	if p.Type == MEM32 {
		return fmt.Sprintf("read%d((void *)0x%08x); // 0x%08x\n", p.AccessSize, p.Address, p.Value)
	} else if p.Type == IO {
		var a string
		if p.AccessSize == 8 {
			a = "b"
		} else if p.AccessSize == 16 {
			a = "w"
		} else if p.AccessSize == 32 {
			a = "l"
		}
		return fmt.Sprintf("in%s(0x%04x); // 0x%08x\n", a, p.Address, p.Value)
	} else if p.Type == MSR {
		return fmt.Sprintf("rdmsr(0x%08x); // 0x%016x\n", p.Address, p.Value)
	} else if p.Type == CPUID {
		return fmt.Sprintf("cpuid(0x%08x); // 0x%016x\n", p.Address, p.Value)
	} else if p.Type == PCI {
		b := (p.Address >> 20) & 0xff
		d := (p.Address >> 15) & 0x1f
		f := (p.Address >> 12) & 0x7
		o := p.Address & 0xfff
		return fmt.Sprintf("pci_read_config%d(PCI_DEV(0x%x, 0x%x, 0x%x), 0x%04x); // 0x%08x\n", p.AccessSize, b, d, f, o, p.Value)
	}
	return ""
}

func (p PWrite) ConvertToC() string {
	if p.Type == MEM32 {
		return fmt.Sprintf("write%d((void *)0x%08x, 0x%08x);\n", p.AccessSize, p.Address, p.Value)
	} else if p.Type == IO {
		var a string
		if p.AccessSize == 8 {
			a = "b"
		} else if p.AccessSize == 16 {
			a = "w"
		} else if p.AccessSize == 32 {
			a = "l"
		}
		return fmt.Sprintf("out%s(0x%x, 0x%04x);\n", a, p.Value, p.Address)
	} else if p.Type == MSR {
		return fmt.Sprintf("{\n\tmsr_t msr = {.lo = 0x%08x, .hi = 0x%08x};\n\twrmsr(0x%08x, msr);\n}\n", p.Value>>32, p.Value&0xffffffff, p.Address)
	} else if p.Type == PCI {
		b := (p.Address >> 20) & 0xff
		d := (p.Address >> 15) & 0x1f
		f := (p.Address >> 12) & 0x7
		o := p.Address & 0xfff

		return fmt.Sprintf("pci_write_config%d(PCI_DEV(0x%x, 0x%x, 0x%x), 0x%04x, 0x%08x);\n", p.AccessSize, b, d, f, o, p.Value)
	}
	return ""
}

func (c CRMW) ConvertToC() string {
	var ret string
	ret += "{\n"

	if c.Type == MEM32 {
		ret += fmt.Sprintf("uint%d_t tmp = read%d((void *)0x%08x);\n", c.AccessSize, c.AccessSize, c.Address)
		ret += fmt.Sprintf("tmp &= ~0x%08x\n", c.AndMask)
		ret += fmt.Sprintf("tmp |= 0x%08x\n", c.OrMask)
		ret += fmt.Sprintf("write%d((void *)0x%08x, tmp);\n", c.AccessSize, c.Address)
	} else if c.Type == IO {
		var a string
		if c.AccessSize == 8 {
			a = "b"
		} else if c.AccessSize == 16 {
			a = "w"
		} else if c.AccessSize == 32 {
			a = "l"
		}
		ret += fmt.Sprintf("uint%d_t tmp = in%s((void *)0x%04x);\n", c.AccessSize, a, c.Address)
		ret += fmt.Sprintf("tmp &= ~0x%08x\n", c.AndMask)
		ret += fmt.Sprintf("tmp |= 0x%08x\n", c.OrMask)
		ret += fmt.Sprintf("out%s((void *)0x%04x, tmp);\n", a, c.Address)
	} else if c.Type == MSR {
		ret += "msr_t msr;\n"
		ret += fmt.Sprintf("msr = rdmsr(0x%08x);\n", c.Address)
		ret += fmt.Sprintf("msr.lo &= ~0x%08x\n", c.AndMask&0xffffffff)
		ret += fmt.Sprintf("msr.hi &= ~0x%08x\n", c.AndMask>>32)
		ret += fmt.Sprintf("msr.lo |=  0x%08x\n", c.OrMask&0xffffffff)
		ret += fmt.Sprintf("msr.hi |=  0x%08x\n", c.OrMask>>32)
		ret += fmt.Sprintf("wrmsr(0x%08x, msr);\n", c.Address)
	} else if c.Type == PCI {
		b := (c.Address >> 20) & 0xff
		d := (c.Address >> 15) & 0x1f
		f := (c.Address >> 12) & 0x7
		o := c.Address & 0xfff

		ret += fmt.Sprintf("uint%d_t tmp = pci_read_config%d(PCI_DEV(0x%x, 0x%x, 0x%x), 0x%04x);\n", c.AccessSize, c.AccessSize, b, d, f, o)
		ret += fmt.Sprintf("tmp &= ~0x%08x\n", c.AndMask)
		ret += fmt.Sprintf("tmp |= 0x%08x\n", c.OrMask)
		ret += fmt.Sprintf("pci_write_config%d(PCI_DEV(0x%x, 0x%x, 0x%x), 0x%04x, tmp);\n", c.AccessSize, b, d, f, o)
	}
	ret += "}\n"
	return ret
}
