package yolocam

import (
	"context"
	"fmt"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

type WhiteBalanceType int

const (
	WhiteBalanceManual WhiteBalanceType = 0
	WhiteBalanceAuto   WhiteBalanceType = 1
)

var whiteBalanceTypes = enum[WhiteBalanceType]{kind: "wb_type", names: map[WhiteBalanceType]string{
	WhiteBalanceManual: "manual",
	WhiteBalanceAuto:   "auto",
}}

func (t WhiteBalanceType) String() string {
	return whiteBalanceTypes.format(t)
}

func (t WhiteBalanceType) MarshalText() ([]byte, error) {
	return whiteBalanceTypes.marshal(t), nil
}

func (t *WhiteBalanceType) UnmarshalText(text []byte) error {
	v, err := whiteBalanceTypes.unmarshal(text)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

func (c *Client) WhiteBalanceType(ctx context.Context) (WhiteBalanceType, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_TYPE, decodeInt[WhiteBalanceType])
}

func (c *Client) SetWhiteBalanceType(ctx context.Context, t WhiteBalanceType) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_TYPE, intValue(t))
}

func (c *Client) WhiteBalanceKelvin(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_KELVIN, decodeInt[int])
}

func (c *Client) SetWhiteBalanceKelvin(ctx context.Context, kelvin int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_KELVIN, intValue(kelvin))
}

func (c *Client) WhiteBalanceTone(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_TONE, decodeInt[int])
}

func (c *Client) SetWhiteBalanceTone(ctx context.Context, tone int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_TONE, intValue(tone))
}

func (c *Client) WhiteBalanceAutoLock(ctx context.Context) (bool, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_AUTO_LOCK, decodeBool)
}

func (c *Client) SetWhiteBalanceAutoLock(ctx context.Context, locked bool) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_AUTO_LOCK, boolValue(locked))
}

type WhiteBalanceOffset struct {
	Red  int `json:"red"`
	Blue int `json:"blue"`
}

func (o WhiteBalanceOffset) String() string {
	return fmt.Sprintf("red=%d,blue=%d", o.Red, o.Blue)
}

func (o WhiteBalanceOffset) value() (*yolocamv1.Value, error) {
	if o.Red < 0 || o.Blue < 0 {
		return nil, fmt.Errorf("offset %v is negative", o)
	}

	return &yolocamv1.Value{Kind: &yolocamv1.Value_WbRedBlueOffset{
		WbRedBlueOffset: &yolocamv1.RedBlueOffset{RedOffset: uint32(o.Red), BlueOffset: uint32(o.Blue)},
	}}, nil
}

func decodeWhiteBalanceOffset(v *yolocamv1.Value) (WhiteBalanceOffset, error) {
	kind, ok := v.GetKind().(*yolocamv1.Value_WbRedBlueOffset)
	if !ok {
		return WhiteBalanceOffset{}, unsupportedValue(v)
	}

	return WhiteBalanceOffset{
		Red:  int(kind.WbRedBlueOffset.GetRedOffset()),
		Blue: int(kind.WbRedBlueOffset.GetBlueOffset()),
	}, nil
}

func (c *Client) WhiteBalanceOffset(ctx context.Context) (WhiteBalanceOffset, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_RED_BLUE_OFFSET, decodeWhiteBalanceOffset)
}

func (c *Client) SetWhiteBalanceOffset(ctx context.Context, o WhiteBalanceOffset) error {
	value, err := o.value()
	if err != nil {
		return propertyError("set", yolocamv1.PropertyId_PROPERTY_ID_WB_RED_BLUE_OFFSET, err)
	}
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_WB_RED_BLUE_OFFSET, value)
}
