package formatter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/deepnoodle-ai/wonton/color"
)

type Cell interface {
	Color() color.Color
	Bold() bool
	String() string
}

type StringCell struct {
	value string
	color *color.Color
	bold  bool
}

func NewStringCell(value string) StringCell {
	return StringCell{
		value: value,
	}
}

func (c StringCell) WithColor(color color.Color) StringCell {
	return StringCell{
		value: c.value,
		color: new(color),
		bold:  c.bold,
	}
}

func (c StringCell) WithBold() StringCell {
	return StringCell{
		value: c.value,
		color: c.color,
		bold:  true,
	}
}

func (c StringCell) Value() string {
	return c.value
}

func (c StringCell) Color() color.Color {
	if c.color == nil {
		return color.NoColor
	}
	return *c.color
}

func (c StringCell) Bold() bool {
	return c.bold
}

func (c StringCell) String() string {
	return c.value
}

type UintCell struct {
	StringCell
	value uint64
}

func NewUIntCell(value uint64) UintCell {
	return UintCell{
		value: value,
	}
}

func (c UintCell) Value() uint64 {
	return c.value
}

func (c UintCell) String() string {
	v := c.StringCell
	v.value = strconv.FormatUint(c.value, 10)
	return v.String()
}

type IntCell struct {
	StringCell
	value int64
}

func NewIntCell(value int64) IntCell {
	return IntCell{
		value: value,
	}
}

func (c IntCell) Value() int64 {
	return c.value
}

func (c IntCell) String() string {
	v := c.StringCell
	v.value = strconv.FormatInt(c.value, 10)
	return v.String()
}

type TimeCell struct {
	StringCell
	value  time.Time
	format string
	empty  string
}

func NewTimeCell(value time.Time) TimeCell {
	return TimeCell{
		value: value,
	}
}

func (c TimeCell) WithFormat(format string) TimeCell {
	c.format = format
	return c
}

func (c TimeCell) WithEmptyValue(value string) TimeCell {
	c.empty = value
	return c
}

func (c TimeCell) Value() time.Time {
	return c.value
}

func (c TimeCell) String() string {
	if c.value.IsZero() || c.value.Equal(time.Unix(0, 0)) {
		return c.empty
	}
	v := c.StringCell
	if c.format == "relative" {
		v.value = formatRelativeTime(c.value)
	} else {
		format := c.format
		if c.format == "" {
			format = "2006-01-02 15:04:05"
		}
		v.value = c.value.Format(format)
	}
	return v.String()
}

type BoolCell struct {
	StringCell
	value bool
}

func NewBoolCell(value bool) BoolCell {
	return BoolCell{
		value: value,
	}
}

func (c BoolCell) Value() bool {
	return c.value
}

func (c BoolCell) String() string {
	v := c.StringCell
	v.value = strconv.FormatBool(c.value)
	return v.String()
}

type SizeCell struct {
	StringCell
	value uint64
}

func NewSizeCell(value uint64) SizeCell {
	return SizeCell{
		value: value,
	}
}

func (c SizeCell) Value() uint64 {
	return c.value
}

func (c SizeCell) String() string {
	v := c.StringCell
	v.value = formatSize(c.value)
	return v.String()
}

func formatSize(size uint64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota) // 1 << 10 = 1024
		MB
		GB
		TB
	)

	var unit string
	var value float64

	switch {
	case size >= TB:
		value = float64(size) / TB
		unit = "TB"
	case size >= GB:
		value = float64(size) / GB
		unit = "GB"
	case size >= MB:
		value = float64(size) / MB
		unit = "MB"
	case size >= KB:
		value = float64(size) / KB
		unit = "kB"
	default:
		value = float64(size)
		unit = "B"
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	now := time.Now()
	diff := now.Sub(t)

	// Handle future dates
	if diff < 0 {
		diff = -diff
		if diff < 24*time.Hour {
			if diff < time.Hour {
				return fmt.Sprintf("in %d minutes", int(diff.Minutes()))
			}
			return fmt.Sprintf("in %d hours", int(diff.Hours()))
		}
		return fmt.Sprintf("in %d days", int(diff.Hours()/24))
	}

	// Handle past dates
	if diff < 24*time.Hour {
		if diff < 2*time.Minute {
			return "now"
		}
		if diff < time.Hour {
			return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		}
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	}

	return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
}
