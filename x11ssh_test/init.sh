#!/bin/bash
IPMIIP="9esec-x11ssh-bmc.9e.network"

ipmitool power off -U admin -P ADMIN -H $IPMIIP
sleep 10
~/SMCIPMITool_2.22.0_build.190701_bundleJRE_Linux_x64/SMCIPMITool $IPMIIP admin ADMIN bios update $(pwd)/coreboot.rom -FORCEREBOOT -f

