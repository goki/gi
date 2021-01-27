// Code generated by "stringer -type=TextDirections"; DO NOT EDIT.

package gist

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LRTB-0]
	_ = x[RLTB-1]
	_ = x[TBRL-2]
	_ = x[LR-3]
	_ = x[RL-4]
	_ = x[TB-5]
	_ = x[LTR-6]
	_ = x[RTL-7]
	_ = x[TextDirectionsN-8]
}

const _TextDirections_name = "LRTBRLTBTBRLLRRLTBLTRRTLTextDirectionsN"

var _TextDirections_index = [...]uint8{0, 4, 8, 12, 14, 16, 18, 21, 24, 39}

func (i TextDirections) String() string {
	if i < 0 || i >= TextDirections(len(_TextDirections_index)-1) {
		return "TextDirections(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TextDirections_name[_TextDirections_index[i]:_TextDirections_index[i+1]]
}

func (i *TextDirections) FromString(s string) error {
	for j := 0; j < len(_TextDirections_index)-1; j++ {
		if s == _TextDirections_name[_TextDirections_index[j]:_TextDirections_index[j+1]] {
			*i = TextDirections(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: TextDirections")
}