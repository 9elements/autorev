#!/bin/bash
IPMIIP="9esec-x11ssh-bmc.9e.network"

ipmitool power off -U admin -P ADMIN -H $IPMIIP
sleep 10
ipmitool power on -U admin -P ADMIN -H $IPMIIP
sleep 1
