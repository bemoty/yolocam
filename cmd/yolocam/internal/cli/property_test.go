package cli

import (
	"context"
	"testing"

	"github.com/bemoty/yolocam"
)

func applySet(p property, raw string) error {
	action, err := p.parseSet(raw)
	if err != nil {
		return err
	}
	return action(context.Background(), nil)
}

func TestAccessor(t *testing.T) {
	var stored yolocam.ExposureType
	p := accessor("mode", "",
		func(*yolocam.Client, context.Context) (yolocam.ExposureType, error) { return stored, nil },
		func(_ *yolocam.Client, _ context.Context, v yolocam.ExposureType) error { stored = v; return nil },
		parseText[yolocam.ExposureType])

	if err := applySet(p, "manual"); err != nil || stored != yolocam.ExposureManual {
		t.Fatalf("set manual = %v, stored %v", err, stored)
	}
	if got, err := p.get(context.Background(), nil); err != nil || got != yolocam.ExposureManual {
		t.Errorf("get = (%v, %v), want manual", got, err)
	}

	_, err := p.parseSet("bogus")
	if err == nil || err.Error() != `mode: invalid exposure_type "bogus", want one of: auto, manual` {
		t.Errorf("bad value err = %v", err)
	}
}
