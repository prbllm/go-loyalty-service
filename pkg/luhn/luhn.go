package luhn

import "strings"

func IsValidOrderNumber(number string) bool {
	number = strings.TrimSpace(number)
	if number == "" {
		return false
	}

	var sum int
	var alt bool

	for i := len(number) - 1; i >= 0; i-- {
		d := number[i]
		if d < '0' || d > '9' {
			return false
		}

		n := int(d - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}

	return sum%10 == 0
}
