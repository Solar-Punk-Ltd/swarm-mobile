# swarm-mobile

SwarmMobile is a bee client built with [fyne](https://fyne.io/) using [bee-lite](https://github.com/Solar-Punk-Ltd/bee-lite). It can run on multiple platforms supported by fyne.

## Build guide

To Build from source you will need **fyne**.

```bash
make get-fyne
```

Also necessary to install the **android ndk** and set the following environment variables.
Then add them to the **PATH** environment variable.

```bash
export ANDROID_HOME=$HOME/Library/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<specific-ndk-version>
export PATH=$ANDROID_HOME:$ANDROID_NDK_HOME:$PATH
```

By default the target is is android/arm64 and the app ID is com.solarpunk.swarmmobile.
On macOS with M chip it TARGET_OS should be darwin
To overwrite them set the following environment variables:

```bash
export APP_ID=com.solarpunk.swarmmobile
export TARGET_OS=darwin
```

To create a package:

```bash
make package
```

## Development

> Recommended to use goenv as go version manager tool

To run without packaging on your local development environment:

```bash
go run main.go
```

If you wish to **simulate a mobile** application:

```bash
go run -tags mobile main.go
```

## Development for Android

Android has networking restrictions since API 30+ so to make [libp2p work](https://github.com/libp2p/go-libp2p/issues/1956) tweaks required on the Go repository that will be used to compile this code (or [bee-lite](https://github.com/Solar-Punk-Ltd/bee-lite/) ).

Based on the following github issues:
[dnsconfig_unix.go](https://github.com/golang/go/issues/8877)
[netlink_linux.go, interface_linux.go](https://github.com/golang/go/issues/40569)

1. Copy the **\_android** files under the **net/** and **syscall/** subfolders of this repo to their respective folders under your go installation, e.g.:

   ```bash
   cp ./net/* /Users/username/.goenv/versions/1.24.2/src/net/
   cp ./syscall/* /Users/username/.goenv/versions/1.24.2/src/syscall/
   ```

2. Furthermore, add the following build directive to the existing dnsconfig_unix, interface_linux, netlink_linux files:

   ```go
   //go:build !android
   ```

   or if a windows exclusion exists add android like

   ```go
   //go:build !windows && !android
   ```

3. To make these changes effect you should recompile the go binaries from the modified sources above - with the newly added \*\_android files .

   For this go to source folder root /Users/username/.goenv/versions/1.24.2/src/ for example.
   Search for make.bash and open it.
   Search for 'bootgo' in this file.
   That version is required to compile the target Go version.
   Install it and point GOROOT_BOOTSTRAP to that GOROOT like GOROOT_BOOTSTRAP=/Users/username/.goenv/versions/1.22.6 on Mac.
   After this run make.bash and it should compile the distro

### Debugging on Android

By building the sources an **.apk** package is generated. It can be installed wiht a simple drag-and-drop on an Android device by connecting your computer via USB. Then just install the package by tapping on the installer (you might need to enable installing packages from unknown sources).

Then on your computer **adb** needs to be installed. Use the following script to start the adb service and listen for the logs coming from the app (it filters out the logs coming from the fyne framework and colors the lines):

```bash
DEVICE_ID=$(adb devices | awk 'FNR == 2 {print $1}')
echo "device ID: ${DEVICE_ID}"
#tail logs
adb -s ${DEVICE_ID} logcat -v color time Fyne:V "*:S"

# or into file
adb -s ${DEVICE_ID} logcat -v color time Fyne:V "*:S" > swarm_mobile.log
```

## Run in the browser

To be able to run the app in the browser set the following environment variable:

```bash
export GOPHERJS_GOROOT=<specific-go-path>/libexec
```

Then run the command:

```bash
fyne serve -os wasm
```

For more information about the build options run:

```bash
fyne package --help
```

## TODO

- [x] release for testnet and mainnet
- [x] code review
- [x] use latest bee-lite version
- [x] fix networking on errors on Android
- [x] host binaries for Android
- [ ] host binaries for IOS
