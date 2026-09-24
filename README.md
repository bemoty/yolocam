# yolocam

This is a reverse-engineering effort of the amazing YoloLiv YoloCam S3 webcam.

Image-quality wise, this is one of the best webcams I've ever had. And it implements UVC, so its video output appears on
pretty much any device completely plug and play. Unfortunately though, its protocol for actually configuring the
camera's image is proprietary and requires the
Electron-based [YoloLiv Compose companion desktop application](https://www.yololiv.com/compose-app) to use.
Linux is also not a supported OS of the Compose app, so Linux users are completely out of luck.

This repository contains:
- The reverse engineered protobuf control protocol in [proto/yolocam/v1/](proto/yolocam/v1/)
- A Go client library for communicating with the camera
- A Go CLI that uses said client library in [cmd/yolocam/](cmd/yolocam/)

## Install

```
go get github.com/bemoty/yolocam
go install github.com/bemoty/yolocam/cmd/yolocam@latest
```

Before you use the CLI, make sure YoloLiv Compose is not running. It seems the webcam doesn't handle two simultaneous
connections to it very well and just bails out until you restart it if you try to connect to it twice.

## Development

```
make build
make check      # fmt-check, vet, staticcheck, buf lint, buf breaking, tests
make generate   # regenerate internal/genproto (needs buf on PATH)
```

`make generate` installs the `protoc-gen-go` version pinned by the tool directive in `go.mod`, so
generated output does not depend on what happens to be installed. Install buf itself separately:
<https://buf.build/docs/installation>.

## Disclaimer

This project is not affiliated with or endorsed by YoloLiv. YoloCam and YoloLiv are trademarks of HongKong Yololiv
Technology Limited. All protocol details, including the protobuf schema in `proto/`, were reconstructed
independently by observing network traffic between the YoloLiv Compose application and my own YoloCam. None of it
is extracted, decompiled, or copied from YoloLiv's software or firmware, and no YoloLiv code, binary, or firmware
is included in this repository.
