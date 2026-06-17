package billing

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type FrozenStore interface {
	Add(ctx context.Context, userID string, delta int64) (int64, error)
	Get(ctx context.Context, userID string) (int64, error)
}

func Reserve(ctx context.Context, store FrozenStore, userID string, balance, estimate int64) error {
	frozen, err := store.Add(ctx, userID, estimate)
	if err != nil {
		return err
	}
	if balance-frozen < 0 {
		_, _ = store.Add(ctx, userID, -estimate)
		return ErrInsufficientBalance
	}
	return nil
}

func Release(ctx context.Context, store FrozenStore, userID string, estimate int64) {
	_, _ = store.Add(ctx, userID, -estimate)
}

func DecimalStringToUnits(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(value, "-")
	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	fraction := "000000"
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) > 6 {
			fraction = fraction[:6]
		}
		for len(fraction) < 6 {
			fraction += "0"
		}
	}
	fractionUnits, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, err
	}
	units := whole*Scale + fractionUnits
	if negative {
		return -units, nil
	}
	return units, nil
}

func UnitsToDecimalString(value int64) string {
	negative := value < 0
	if negative {
		value = -value
	}
	whole := value / Scale
	fraction := value % Scale
	if negative {
		return "-" + strconv.FormatInt(whole, 10) + "." + leftPad6(fraction)
	}
	return strconv.FormatInt(whole, 10) + "." + leftPad6(fraction)
}

func leftPad6(value int64) string {
	raw := strconv.FormatInt(value, 10)
	for len(raw) < 6 {
		raw = "0" + raw
	}
	return raw
}
