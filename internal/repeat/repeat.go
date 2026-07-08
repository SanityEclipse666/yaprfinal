package repeat

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat rule is empty")
	}

	// Парсим начальную дату в локальном времени
	startDate, err := time.ParseInLocation(DateLayout, dstart, now.Location())
	if err != nil {
		return "", fmt.Errorf("invalod dstart format: %v", err)
	}

	// Сбрасываем время у now до полуночи в локальном поясе
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	parts := strings.Split(repeat, " ")
	ruleType := parts[0]

	var nextDate time.Time

	switch ruleType {
	case "y":
		nextDate = startDate
		for {
			nextDate = nextDate.AddDate(1, 0, 0)

			if nextDate.After(now) {
				break
			}
		}

	case "d":
		if len(parts) < 2 {
			return "", errors.New("no interval for 'd'")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.New("incorrect interval for 'd'")
		}
		if days > 400 {
			return "", errors.New("interval can't be more than 400 days")
		}

		nextDate = startDate
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(now) {
				break
			}
		}

	default:
		return "", fmt.Errorf("format is unsupported: %s", ruleType)
	}

	return nextDate.Format(DateLayout), nil
}
