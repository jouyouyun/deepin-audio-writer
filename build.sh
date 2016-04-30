#!/bin/bash

sudo mv /usr/share/gocode/src/pkg.linuxdeepin.com/lib/pulse /tmp/pulse.bak
sudo cp -R pulse /usr/share/gocode/src/pkg.linuxdeepin.com/lib/
GOPATH=/usr/share/gocode; go build -o deepin-audio-writer audio_writer.go
sudo rm -rf /usr/share/gocode/src/pkg.linuxdeepin.com/lib/pulse
sudo mv /tmp/pulse.bak /usr/share/gocode/src/pkg.linuxdeepin.com/lib/
