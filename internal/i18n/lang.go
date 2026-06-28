package i18n

import "strings"

type Lang string

const (
	EN Lang = "en"
	RU Lang = "ru"
)

func Parse(s string) Lang {
	switch Lang(strings.ToLower(strings.TrimSpace(s))) {
	case EN:
		return EN
	case RU:
		return RU
	default:
		return RU
	}
}

func (l Lang) Toggle() Lang {
	if l == EN {
		return RU
	}
	return EN
}
