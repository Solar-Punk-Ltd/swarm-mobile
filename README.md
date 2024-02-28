# swarm-mobile

SwarmMobile is a bee client built with [fyne](https://fyne.io/) using [bee-lite](https://github.com/onepeerlabs/bee-lite). It can run on multiple platforms supported by fyne.

## Build guide

To Build from source you will need **fyne**.
```
make get-fnye
```

Also necessary to install the **android ndk** and set the following environment variables.
Then add them to the **PATH** environment variable.
```
export ANDROID_HOME=$HOME/Library/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<specific-ndk-version>
export PATH=$ANDROID_HOME:$ANDROID_NDK_HOME:$PATH
```

By default the target is is android/arm64 and the app ID is com.solarpunk.swarmmobile.
To overwrite them set the following environment variables:
```
export APP_ID=<app-id>
export TARGET_OS=<target-os>
```

To create a package:
```
make package
```

## Development

To run without packaging on your local development environment:
```
go run main.go
```

If you wish to **simulate a mobile** application:
```
go run -tags mobile main.go
```

In order for the android networking to work:
Copy the ***_android** files under the **net/** and **syscall/** subfolders of this repo to their respective folders under your go installation, e.g.:
```
cp ./net/* /opt/homebrew/Cellar/go/1.22.0/libexec/src/net/
cp ./syscall/* /opt/homebrew/Cellar/go/1.22.0/libexec/src/syscall/
```
Based on the following github issues:
[dnsconfig_unix.go](https://github.com/golang/go/issues/8877)
[netlink_linux.go, interface_linux.go](https://github.com/golang/go/issues/40569)

## Run in the browser

To be able to run the app in the browser set the following environment variable:
```
export GOPHERJS_GOROOT=<specific-go-path>/libexec
```

Then run the command:
```
fyne serve -os wasm
```

For more information about the build options run:
```
fyne package --help
```

## TODO
- [ ] release for testnet and mainnet
- [ ] host binaries
- [ ] code review
