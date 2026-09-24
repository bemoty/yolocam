/*
Package yolocam provides methods to control settings of YoloLiv's YoloCam webcams.

The primary use case of this library is getting and setting the webcam's properties such as exposure, white balance,
sharpness, etc. It does not offer APIs for actually reading video off the webcam, as this is already accessible
through the cam's native UVC support.

Use [Connect] to get a [Client] for interacting with the webcam:

	client, err := yolocam.Connect(ctx, nil)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.SetExposureType(ctx, yolocam.ExposureManual); err != nil {
		return err
	}
	if err := client.SetExposureISO(ctx, 400); err != nil {
		return err
	}

Passing nil uses the webcam's default address. See [Options] if yours differs for whatever reason.

# Only one at a time!

It seems the webcam is not very happy if you try to open more than one connection to it, as it will simply stop
responding to any requests until you restart it. So make sure to close YoloLiv Compose before connecting and do not
open more than one [Client] per webcam.

# Watching the webcam

While streaming video, the webcam likes to send unsolicited messages of its own. It's mostly telemetry, but if you
set the cam's focus type to face tracking, the webcam will send messages that contain information about where it
detected a face several times a second.

You can use [Client.Watch] to subscribe to these messages.

# Errors

All getters and setters return a [*PropertyError], which tells you which property failed and whether it was a get or a
set. What actually went wrong is wrapped inside, so use [errors.Is] and [errors.As] to get to it:

  - [ErrNotImplemented] if the webcam doesn't know the property
  - [*StatusError] if the webcam refuses the request with some status tha the library does not know
  - [*UnsupportedValueError] if the request to the webcam went through, but the replied value is unexpected (e.g., text
    where a number should be)
  - [ErrClosed] if the [Client] was already closed
  - Context errors if ctx was canceled or the request timed out

If YoloLiv ever changes a property in a firmware update, you will most likely run into one of the first three.

# Firmware

So far, this library has only been tested against firmware 1.0.0 (build 1295), since that's what my own webcam runs.
Your mileage may vary for any other firmware version. You can use [DeviceInfo.FirmwareVerified] to find out whether
your webcam runs a tested version.
*/
package yolocam
