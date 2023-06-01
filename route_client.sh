#!/bin/bash
ifconfig tun0 13.0.0.2 netmask 255.255.255.255
ip route add 13.0.0.1 dev tun0
