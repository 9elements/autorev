# QEMU test environment

The scripts in this folder allow to start/stop and restart qemu,
running a precompiled coreboot image.

The coreboot image contains a autorev shell and communicates
with the guest serial over two fifos.


## Tutorial

Set up a mysql server. Then run:

	mysql -u root -p autorev < autorev_withoutData.sq

Update the config.yml and add the connection details to your mysql server:

`
database:
        hostname: localhost
        port: 3306
        username: root
        password: YourPasswordH3r3
`

Add the default BIOS option config binary:

	./autorev -newConfig -newConfigFile qemu_test/qemuTestDefaults.bin -newConfigName "qemuConfig"

Add new tracelog for every possible BIOS option combination as defined in config.yml:

	./autorev -newtrace

Run the newly created tracelogs on qemu and gather data:

	./autorev -collecttraces

## coreboot.rom

coreboot.rom is build from coreboot.org 26060bc7c852b6b6ecf6ecb0e225d4ef414c8f6f
+ Change-Id: I13e47f45e69376d046f35c04363fe3db1cfaa610
+ https://github.com/PatrickRudolph/libx86emu/tree/cpuid_msr_callback

## config.yaml

An example configuration to use qemu as debug target.

## Example BLOB

The coreboot.rom contains an example blob. The source code is
```
        outw(0xddaa, 0x80);

        if (config) {
                struct testconfig {
                        uint32_t config1;
                        uint32_t config2;
                        uint32_t config3;
                        uint32_t config4;
                } *myconfig = (struct testconfig *)config;

                if (myconfig->config1 || myconfig->config2) {
                        pci_write_config8(dev, Q35_PAM0 + 6, 0x30);
                }
                if (myconfig->config1) {
                        pci_read_config32(dev, Q35_PAM0);
                }
                if (myconfig->config2) {
                        outb(0x00, 0x80);
                }
                pci_write_config8(dev, Q35_PAM0 + 2, myconfig->config3);
                if (myconfig->config4) {
                        pci_write_config8(dev, Q35_PAM0 + 1, myconfig->config4 + 1);
                } else {
                        outb(0x12, 0x80);
                        outb(0x34, 0x80);
                }
        }
        outw(0xaadd, 0x80);
```

