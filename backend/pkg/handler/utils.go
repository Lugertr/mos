package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrUserNotFound = fmt.Errorf("user id not found in context")

// parseDateFlexible пытается распарсить строку в time.Time.
// Поддерживаемые форматы (попытки в порядке):
// - RFC3339 / time.RFC3339
// - "2006-01-02 15:04:05"
// - "2006-01-02"
// - unix seconds (целое число) и unix milliseconds (13 цифр)
func parseDateFlexible(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty date string")
	}

	// Попробуем RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Попробуем разобрать как unix timestamp (секунды или миллисекунды)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		// 13+ цифр — миллисекунды
		if len(s) >= 13 {
			return time.Unix(0, i*int64(time.Millisecond)).UTC(), nil
		}
		// иначе секунды
		return time.Unix(i, 0).UTC(), nil
	}

	return time.Time{}, errors.New("unsupported date format")
}
