package service

import "lingshu/backend/internal/billing"

func decimalMul(left, right string) string {
	leftUnits, err := billing.DecimalStringToUnits(left)
	if err != nil {
		return "0.000000"
	}
	rightUnits, err := billing.DecimalStringToUnits(right)
	if err != nil {
		return "0.000000"
	}
	return billing.UnitsToDecimalString((leftUnits * rightUnits) / billing.Scale)
}

func unitsToDecimalString(value int64) string {
	return billing.UnitsToDecimalString(value)
}
