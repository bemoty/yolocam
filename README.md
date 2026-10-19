# yolocam

![yolocam setting options while the camera feed updates live](.github/assets/demo.avif)

This is a reverse-engineering effort of the amazing YoloLiv YoloCam S3 webcam (and probably other YoloLiv webcams).

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

## Usage

If you want to use this in your Go app, simply import the library and keep one `Client` around for as long as you need 
the webcam. You can use the CLI `set` and `get` commands for one-off commands to the camera, or use 
`yolocam serve --stdio` as a child process and keep it running. It holds one connection to the webcam
and speaks [JSON-RPC 2.0](https://www.jsonrpc.org/specification) over its stdin and stdout, one JSON message per
line.

```
$ yolocam serve --stdio
{"jsonrpc":"2.0","id":1,"method":"set","params":{"property":"wb_kelvin","value":5600}}
{"jsonrpc":"2.0","result":null,"id":1}
```

`yolocam openrpc` prints an [OpenRPC](https://spec.open-rpc.org/) document with everything else you could need: methods,
params, value schemas, errors, exit codes and push notifications.

## Development

You will need [GNU Make](https://www.gnu.org/software/make/), [Go](https://go.dev/), and [buf](https://buf.build/).

```
make build
make check      # fmt-check, vet, staticcheck, buf lint, buf breaking, tests
make generate   # regenerate internal/genproto (needs buf on PATH)
```

## Disclaimer

This project is not affiliated with or endorsed by YoloLiv. YoloCam and YoloLiv are trademarks of HongKong Yololiv
Technology Limited. All protocol details, including the protobuf schema in `proto/`, were reconstructed
independently by observing network traffic between the YoloLiv Compose application and my own YoloCam. None of it
is extracted, decompiled, or copied from YoloLiv's software or firmware, and no YoloLiv code, binary, or firmware
is included in this repository.
