package yolocam

import (
	"context"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

func (c *Client) ExposureISO(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO, decodeInt[int])
}

func (c *Client) SetExposureISO(ctx context.Context, iso int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO, intValue(iso))
}

type ExposureType int

const (
	ExposureAuto   ExposureType = 0
	ExposureManual ExposureType = 1
)

var exposureTypes = enum[ExposureType]{kind: "exposure_type", names: map[ExposureType]string{
	ExposureAuto:   "auto",
	ExposureManual: "manual",
}}

func (t ExposureType) String() string {
	return exposureTypes.format(t)
}

func (t ExposureType) MarshalText() ([]byte, error) {
	return exposureTypes.marshal(t), nil
}

func (t *ExposureType) UnmarshalText(text []byte) error {
	v, err := exposureTypes.unmarshal(text)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

func (c *Client) ExposureType(ctx context.Context) (ExposureType, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE, decodeInt[ExposureType])
}

func (c *Client) SetExposureType(ctx context.Context, t ExposureType) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE, intValue(t))
}
