package entity

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type LuckyDrawPercentage int64

var luckyDrawPercentagePattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,6})(\.[0-9]{1,2})?$`)

func (p *LuckyDrawPercentage) UnmarshalJSON(data []byte) error {
	value := string(bytes.TrimSpace(data))
	if !luckyDrawPercentagePattern.MatchString(value) {
		return fmt.Errorf("percentage must be a JSON number with at most two decimal places")
	}
	negative := strings.HasPrefix(value, "-")
	parts := strings.Split(strings.TrimPrefix(value, "-"), ".")
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	var fraction int64
	if len(parts) == 2 {
		fraction, err = strconv.ParseInt(parts[1]+strings.Repeat("0", 2-len(parts[1])), 10, 64)
		if err != nil {
			return err
		}
	}
	units := whole*100 + fraction
	if negative {
		units = -units
	}
	*p = LuckyDrawPercentage(units)
	return nil
}

func (p LuckyDrawPercentage) MarshalJSON() ([]byte, error) {
	units := int64(p)
	if units < 0 || units > 10000 {
		return nil, fmt.Errorf("stored percentage is out of range")
	}
	return []byte(fmt.Sprintf("%d.%02d", units/100, units%100)), nil
}
