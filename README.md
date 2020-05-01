# OSMC Remote Controller (OSMC / Remote PIS-0445) on Raspberry PI on Linux for Google Chromecast

A Go example library for using OSMC remote control on Raspberry PI on Linux that can control
a Google Chromecast.


## Building for Raspberry PI

	GOOS=linux GOARCH=arm GOARM=7 go build -o ./build/armv7-linux/osmc-controller

## Remote Controller Functions

```
- ARROW_UP -> volume up
- ARROW_DOWN -> volume down
- ARROW_LEFT -> rewind
- ARROW_RIGHT -> fast-forward
- OK -> mute
- PLAY_PAUSE -> play / pause
- STOP -> stop playing media
- PREV -> play previous track (or go back to the start)
- NEXT -> skip to next track
```
