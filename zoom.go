// Copyright 2026 Joshua Winkler and The yolocam Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yolocam

import (
	"context"
	"fmt"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

const (
	MinZoomFactor = 1.0
	MaxZoomFactor = 4.0
)

// Zoom is a digital zoom into the webcam's image. Factor ranges from [MinZoomFactor] to [MaxZoomFactor]. CenterX and
// CenterY place the zoomed area within the unmirrored image, from 0 (left, top) to 1 (right, bottom).
type Zoom struct {
	Factor  float64 `json:"factor"`
	CenterX float64 `json:"center_x"`
	CenterY float64 `json:"center_y"`
}

// CenteredZoom zooms into the middle of the image.
func CenteredZoom(factor float64) Zoom {
	return Zoom{Factor: factor, CenterX: 0.5, CenterY: 0.5}
}

func (z Zoom) String() string {
	return fmt.Sprintf("%g@%.2f,%.2f", z.Factor, z.CenterX, z.CenterY)
}

// value clamps the center so the zoomed area stays inside the image, the same way Compose limits its Location
// control.
func (z Zoom) value() (*yolocamv1.Value, error) {
	if z.Factor < MinZoomFactor || z.Factor > MaxZoomFactor {
		return nil, fmt.Errorf("zoom factor %g is outside %g to %g", z.Factor, MinZoomFactor, MaxZoomFactor)
	}

	half := 0.5 / z.Factor
	cx := min(max(z.CenterX, half), 1-half)
	cy := min(max(z.CenterY, half), 1-half)
	return &yolocamv1.Value{Kind: &yolocamv1.Value_ZoomRect{ZoomRect: &yolocamv1.ZoomRect{
		X1:         float32(cx - half),
		Y1:         float32(cy - half),
		ZoomFactor: float32(z.Factor),
		X2:         float32(cx + half),
		Y2:         float32(cy + half),
	}}}, nil
}

// Zoom reads the value set by [Client.SetZoom].
func (c *Client) Zoom(ctx context.Context) (Zoom, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_ZOOM, decodeZoom)
}

// SetZoom zooms into the image. A factor outside [MinZoomFactor] to [MaxZoomFactor] is rejected before anything is
// sent, and a center too close to the edge is moved inwards until the zoomed area fits.
func (c *Client) SetZoom(ctx context.Context, z Zoom) error {
	value, err := z.value()
	if err != nil {
		return propertyError("set", yolocamv1.PropertyId_PROPERTY_ID_ZOOM, err)
	}
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_ZOOM, value)
}

func decodeZoom(v *yolocamv1.Value) (Zoom, error) {
	kind, ok := v.GetKind().(*yolocamv1.Value_ZoomRect)
	if !ok {
		return Zoom{}, unsupportedValue(v)
	}

	r := kind.ZoomRect
	return Zoom{
		Factor:  float64(r.GetZoomFactor()),
		CenterX: float64(r.GetX1()+r.GetX2()) / 2,
		CenterY: float64(r.GetY1()+r.GetY2()) / 2,
	}, nil
}
