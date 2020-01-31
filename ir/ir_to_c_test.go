package ir

import "testing"

func TestPRead_ConvertToC(t *testing.T) {
	type fields struct {
		Type       PrimtiveType
		Address    uint
		Value      uint64
		AccessSize uint
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"read8",
			fields{MEM32, 0, 0x1234, 8},
			"read8((void *)0x00000000); // 0x00001234\n",
		},
		{
			"read16",
			fields{MEM32, 1, 0x12345, 16},
			"read16((void *)0x00000001); // 0x00012345\n",
		},
		{
			"read32",
			fields{MEM32, 2, 0x123456, 32},
			"read32((void *)0x00000002); // 0x00123456\n",
		},
		{
			"inb",
			fields{IO, 0, 0xaa, 8},
			"inb(0x0000); // 0x000000aa\n",
		},
		{
			"inw",
			fields{IO, 1, 0xaabb, 16},
			"inw(0x0001); // 0x0000aabb\n",
		},
		{
			"inl",
			fields{IO, 2, 0x11223344, 32},
			"inl(0x0002); // 0x11223344\n",
		},
		{
			"msr",
			fields{MSR, 0x67, 0xaa | 0xbb<<32, 8},
			"rdmsr(0x00000067); // 0x000000bb000000aa\n",
		},
		{
			"msr",
			fields{MSR, 0x67, 0xbb | 0xaa<<32, 16},
			"rdmsr(0x00000067); // 0x000000aa000000bb\n",
		},
		{
			"msr",
			fields{MSR, 0x67, 0, 32},
			"rdmsr(0x00000067); // 0x0000000000000000\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PRead{
				Type:       tt.fields.Type,
				Address:    tt.fields.Address,
				Value:      tt.fields.Value,
				AccessSize: tt.fields.AccessSize,
			}
			if got := p.ConvertToC(); got != tt.want {
				t.Errorf("PRead.ConvertToC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPWrite_ConvertToC(t *testing.T) {
	type fields struct {
		Type       PrimtiveType
		Address    uint
		Value      uint64
		AccessSize uint
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"write8",
			fields{MEM32, 0, 0x1234, 8},
			"write8((void *)0x00000000, 0x00001234);\n",
		},
		{
			"write16",
			fields{MEM32, 1, 0x12345, 16},
			"write16((void *)0x00000001, 0x00012345);\n",
		},
		{
			"write32",
			fields{MEM32, 2, 0x123456, 32},
			"write32((void *)0x00000002, 0x00123456);\n",
		},
		{
			"outb",
			fields{IO, 0, 0xaa, 8},
			"outb(0xaa, 0x0000);\n",
		},
		{
			"outw",
			fields{IO, 1, 0xaabb, 16},
			"outw(0xaabb, 0x0001);\n",
		},
		{
			"outl",
			fields{IO, 2, 0x11223344, 32},
			"outl(0x11223344, 0x0002);\n",
		},
		{
			"wrmsr",
			fields{MSR, 0x67, 0xaa | 0xbb<<32, 8},
			"{\n\tmsr_t msr = {.lo = 0x000000bb, .hi = 0x000000aa};\n\twrmsr(0x00000067, msr);\n}\n",
		},
		{
			"wrmsr",
			fields{MSR, 0x67, 0xbb | 0xaa<<32, 16},
			"{\n\tmsr_t msr = {.lo = 0x000000aa, .hi = 0x000000bb};\n\twrmsr(0x00000067, msr);\n}\n",
		},
		{
			"wrmsr",
			fields{MSR, 0x67, 0, 32},
			"{\n\tmsr_t msr = {.lo = 0x00000000, .hi = 0x00000000};\n\twrmsr(0x00000067, msr);\n}\n",
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PWrite{
				Type:       tt.fields.Type,
				Address:    tt.fields.Address,
				Value:      tt.fields.Value,
				AccessSize: tt.fields.AccessSize,
			}
			if got := p.ConvertToC(); got != tt.want {
				t.Errorf("PWrite.ConvertToC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCRMW_ConvertToC(t *testing.T) {
	type fields struct {
		Type       PrimtiveType
		Address    uint
		OrMask     uint64
		AndMask    uint64
		AccessSize uint
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CRMW{
				Type:       tt.fields.Type,
				Address:    tt.fields.Address,
				OrMask:     tt.fields.OrMask,
				AndMask:    tt.fields.AndMask,
				AccessSize: tt.fields.AccessSize,
			}
			if got := c.ConvertToC(); got != tt.want {
				t.Errorf("CRMW.ConvertToC() = %v, want %v", got, tt.want)
			}
		})
	}
}
