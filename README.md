# swarm-mobile

SwarmMobile is a bee client built with [fyne](https://fyne.io/) using [bee-lite](https://github.com/onepeerlabs/bee-lite). It can run on multiple platforms supported by fyne.

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
To overwrite them set the following environment variables:
```bash
export APP_ID=<app-id>
export TARGET_OS=<target-os>
```

To create a package:
```bash
make package
```

## Development

To run without packaging on your local development environment:
```bash
go run main.go
```

If you wish to **simulate a mobile** application:
```bash
go run -tags mobile main.go
```

In order for the android networking to work:
Copy the **_android** files under the **net/** and **syscall/** subfolders of this repo to their respective folders under your go installation, e.g.:
```bash
cp ./net/* /opt/homebrew/Cellar/go/1.22.0/libexec/src/net/
cp ./syscall/* /opt/homebrew/Cellar/go/1.22.0/libexec/src/syscall/
```
Based on the following github issues:
[dnsconfig_unix.go](https://github.com/golang/go/issues/8877)
[netlink_linux.go, interface_linux.go](https://github.com/golang/go/issues/40569)

### Debugging on Android

By building the sources an **.apk** package is generated. It can be installed wiht a simple drag-and-drop on an Android device by connecting your computer via USB. Then just install the package by tapping on the installer (you might need to enable installing packages from unknown sources).

Then on your computer **adb** needs to be installed. Use the following script to start the adb service and listen for the logs coming from the app (it filters out the logs coming from the fyne framework and colors the lines):
```bash
DEVICE_ID=$(adb devices | awk 'FNR == 2 {print $1}')
echo "device ID: ${DEVICE_ID}"
adb logcat -v color time Fyne:V *:S ${DEVICE_ID} > swarm_mobile.log
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
- [ ] release for testnet and mainnet
- [ ] host binaries
- [ ] code review
