// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gist

import (
	"fmt"
	"log"
	"strings"

	"github.com/goki/gi/units"
	"github.com/goki/mat32"
)

// Sides contains values for each side or corner of a box.
// If Sides contains sides, the struct field names correspond
// directly to the side values (ie: Top = top side value).
// If Sides contains corners, the struct field names correspond
// to the corners as follows: Top = top left, Right = top right,
// Bottom = bottom right, Left = bottom left.
type Sides[T comparable] struct {
	Top    T `xml:"top" desc:"top/top-left value"`
	Right  T `xml:"right" desc:"right/top-right value"`
	Bottom T `xml:"bottom" desc:"bottom/bottom-right value"`
	Left   T `xml:"left" desc:"left/bottom-left value"`
}

// NewSides is a helper that creates new sides/corners of the given type
// and calls Set on them with the given values.
// It does not return any error values and just logs them.
func NewSides[T comparable](vals ...T) Sides[T] {
	sides, _ := NewSidesTry[T](vals...)
	return sides
}

// NewSidesTry is a helper that creates new sides/corners of the given type
// and calls Set on them with the given values.
// It returns an error value if there is one.
func NewSidesTry[T comparable](vals ...T) (Sides[T], error) {
	sides := Sides[T]{}
	err := sides.Set(vals...)
	return sides, err
}

// Set sets the values of the sides/corners from the given list of 0 to 4 values.
// If 0 values are provided, all sides/corners are set to the zero value of the type.
// If 1 value is provided, all sides/corners are set to that value.
// If 2 values are provided, the top/top-left and bottom/bottom-right are set to the first value
// and the right/top-right and left/bottom-left are set to the second value.
// If 3 values are provided, the top/top-left is set to the first value,
// the right/top-right and left/bottom-left are set to the second value,
// and the bottom/bottom-right is set to the third value.
// If 4 values are provided, the top/top-left is set to the first value,
// the right/top-right is set to the second value, the bottom/bottom-right is set
// to the third value, and the left/bottom-left is set to the fourth value.
// If more than 4 values are provided, the behavior is the same
// as with 4 values, but Set also prints and returns
// an error. This error is not critical and does not need to be
// handled, as the values are still set, but it can be if wished.
// This behavior is based on the CSS multi-side/corner setting syntax,
// like that with padding and border-radius (see https://www.w3schools.com/css/css_padding.asp
// and https://www.w3schools.com/cssref/css3_pr_border-radius.php)
func (s *Sides[T]) Set(vals ...T) error {
	switch len(vals) {
	case 0:
		var zval T
		s.SetAll(zval)
	case 1:
		s.SetAll(vals[0])
	case 2:
		s.SetVert(vals[0])
		s.SetHoriz(vals[1])
	case 3:
		s.Top = vals[0]
		s.SetHoriz(vals[1])
		s.Bottom = vals[2]
	case 4:
		s.Top = vals[0]
		s.Right = vals[1]
		s.Bottom = vals[2]
		s.Left = vals[3]
	default:
		s.Top = vals[0]
		s.Right = vals[1]
		s.Bottom = vals[2]
		s.Left = vals[3]
		err := fmt.Errorf("sides.Set: expected 0 to 4 values, but got %d", len(vals))
		log.Println(err)
		return err
	}
	return nil
}

// SetVert sets the values for the sides/corners in the
// vertical/diagonally descending direction
// (top/top-left and bottom/bottom-right) to the given value
func (s *Sides[T]) SetVert(val T) {
	s.Top = val
	s.Bottom = val
}

// SetHoriz sets the values for the sides/corners in the
// horizontal/diagonally ascending direction
// (right/top-right and left/bottom-left) to the given value
func (s *Sides[T]) SetHoriz(val T) {
	s.Right = val
	s.Left = val
}

// SetAll sets the values for all of the sides/corners
// to the given value
func (s *Sides[T]) SetAll(val T) {
	s.Top = val
	s.Right = val
	s.Bottom = val
	s.Left = val
}

// This returns the sides/corners as a Sides value
// (instead of some higher-level value in which
// the sides/corners are embedded)
func (s Sides[T]) This() Sides[T] {
	return s
}

// AllSame returns whether all of the sides/corners are the same
func (s Sides[T]) AllSame() bool {
	return s.Right == s.Top && s.Bottom == s.Top && s.Left == s.Top
}

// IsZero returns whether all of the sides/corners are equal to zero
func (s Sides[T]) IsZero() bool {
	var zval T
	return s.Top == zval && s.Right == zval && s.Bottom == zval && s.Left == zval
}

// SetStringer is a type that can be set from a string
type SetStringer interface {
	SetString(str string) error
}

// SetAny sets the sides/corners from the given value of any type
func (s *Sides[T]) SetAny(a any) error {
	switch val := a.(type) {
	case Sides[T]:
		*s = val
	case *Sides[T]:
		*s = *val
	case T:
		s.SetAll(val)
	case *T:
		s.SetAll(*val)
	case []T:
		s.Set(val...)
	case *[]T:
		s.Set(*val...)
	case string:
		return s.SetString(val)
	default:
		return s.SetString(fmt.Sprint(val))
	}
	return nil
}

// SetString sets the sides/corners from the given string value
func (s *Sides[T]) SetString(str string) error {
	fields := strings.Fields(str)
	vals := make([]T, len(fields))
	for i, field := range fields {
		ss, ok := any(&vals[i]).(SetStringer)
		if !ok {
			err := fmt.Errorf("(Sides).SetString('%s'): to set from a string, the sides type (%T) must implement SetStringer (needs SetString(str string) error function)", str, s)
			log.Println(err)
			return err
		}
		err := ss.SetString(field)
		if err != nil {
			nerr := fmt.Errorf("(Sides).SetString('%s'): error setting sides of type %T from string: %w", str, s, err)
			log.Println(nerr)
			return nerr
		}
	}
	return s.Set(vals...)
}

// SideValues contains units.Value values for each side/corner of a box
type SideValues struct {
	Sides[units.Value]
}

// NewSideValues is a helper that creates new side/corner values
// and calls Set on them with the given values.
// It does not return any error values and just logs them.
func NewSideValues(vals ...units.Value) SideValues {
	sides, _ := NewSideValuesTry(vals...)
	return sides
}

// NewSideValuesTry is a helper that creates new side/corner
// values and calls Set on them with the given values.
// It returns an error value if there is one.
func NewSideValuesTry(vals ...units.Value) (SideValues, error) {
	sides := Sides[units.Value]{}
	err := sides.Set(vals...)
	return SideValues{Sides: sides}, err
}

// ToDots converts the values for each of the sides/corners
// to raw display pixels (dots) and sets the Dots field for each
// of the values. It returns the dot values as a SideFloats.
func (sv *SideValues) ToDots(uc *units.Context) SideFloats {
	return NewSideFloats(
		sv.Top.ToDots(uc),
		sv.Right.ToDots(uc),
		sv.Bottom.ToDots(uc),
		sv.Left.ToDots(uc),
	)
}

// Dots returns the dot values of the sides/corners as a SideFloats.
// It does not compute them; see ToDots for that.
func (sv SideValues) Dots() SideFloats {
	return NewSideFloats(
		sv.Top.Dots,
		sv.Right.Dots,
		sv.Bottom.Dots,
		sv.Left.Dots,
	)
}

// SideFloats contains float32 values for each side/corner of a box
type SideFloats struct {
	Sides[float32]
}

// NewSideFloats is a helper that creates new side/corner floats
// and calls Set on them with the given values.
// It does not return any error values and just logs them.
func NewSideFloats(vals ...float32) SideFloats {
	sides, _ := NewSideFloatsTry(vals...)
	return sides
}

// NewSideFloatsTry is a helper that creates new side/corner floats
// and calls Set on them with the given values.
// It returns an error value if there is one.
func NewSideFloatsTry(vals ...float32) (SideFloats, error) {
	sides := Sides[float32]{}
	err := sides.Set(vals...)
	return SideFloats{Sides: sides}, err
}

// Pos returns the position offset casued by the side/corner values (Left, Top)
func (sf SideFloats) Pos() mat32.Vec2 {
	return mat32.NewVec2(sf.Left, sf.Top)
}

// Size returns the toal size the side/corner values take up (Left + Right, Top + Bottom)
func (sf SideFloats) Size() mat32.Vec2 {
	return mat32.NewVec2(sf.Left+sf.Right, sf.Top+sf.Bottom)
}

// SideColors contains color values for each side/corner of a box
type SideColors struct {
	Sides[Color]
}

// NewSideColors is a helper that creates new side/corner colors
// and calls Set on them with the given values.
// It does not return any error values and just logs them.
func NewSideColors(vals ...Color) SideColors {
	sides, _ := NewSideColorsTry(vals...)
	return sides
}

// NewSideColorsTry is a helper that creates new side/corner colors
// and calls Set on them with the given values.
// It returns an error value if there is one.
func NewSideColorsTry(vals ...Color) (SideColors, error) {
	sides := Sides[Color]{}
	err := sides.Set(vals...)
	return SideColors{Sides: sides}, err
}

// SetAny sets the sides/corners from the given value of any type
func (s *SideColors) SetAny(a any, ctxt Context) error {
	switch val := a.(type) {
	case Sides[Color]:
		s.Sides = val
	case *Sides[Color]:
		s.Sides = *val
	case Color:
		s.SetAll(val)
	case *Color:
		s.SetAll(*val)
	case []Color:
		s.Set(val...)
	case *[]Color:
		s.Set(*val...)
	case string:
		return s.SetString(val, ctxt)
	default:
		return s.SetString(fmt.Sprint(val), ctxt)
	}
	return nil
}

// SetString sets the sides/corners from the given string value
func (s *SideColors) SetString(str string, ctxt Context) error {
	fields := strings.Fields(str)
	vals := make([]Color, len(fields))
	for i, field := range fields {
		err := (&vals[i]).SetStringStyle(field, nil, ctxt)
		if err != nil {
			nerr := fmt.Errorf("(SideColors).SetString('%s'): error setting sides of type %T from string: %w", str, s, err)
			log.Println(nerr)
			return nerr
		}
	}
	return s.Set(vals...)
}