#!/bin/bash
IPMIIP="9esec-x11ssh-bmc.9e.network"

ipmitool power reset -U admin -P ADMIN -H $IPMIIP

