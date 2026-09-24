package cli

import (
	"context"
	"fmt"

	"github.com/bemoty/yolocam"
)

var exposureProperties = []property{
	accessor("exposure", "ISO sensitivity, 100 to 6400 in third stops",
		(*yolocam.Client).ExposureISO, (*yolocam.Client).SetExposureISO, parseInt,
	).withWarning(warnIfAutoExposure),
	accessor("exposure-type", "auto or manual exposure",
		(*yolocam.Client).ExposureType, (*yolocam.Client).SetExposureType, parseText),
}

func warnIfAutoExposure(ctx context.Context, c *yolocam.Client) string {
	mode, err := c.ExposureType(ctx)
	if err != nil || mode != yolocam.ExposureAuto {
		return ""
	}
	return fmt.Sprintf("no visible change while exposure_type is %s", mode)
}
