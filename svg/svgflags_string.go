// Code generated by "stringer -type=SVGFlags"; DO NOT EDIT.

package svg

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Rendering-35]
	_ = x[SVGFlagsN-36]
}

const _SVGFlags_name = "RenderingSVGFlagsN"

var _SVGFlags_index = [...]uint8{0, 9, 18}

func (i SVGFlags) String() string {
	i -= 35
	if i < 0 || i >= SVGFlags(len(_SVGFlags_index)-1) {
		return "SVGFlags(" + strconv.FormatInt(int64(i+35), 10) + ")"
	}
	return _SVGFlags_name[_SVGFlags_index[i]:_SVGFlags_index[i+1]]
}

func StringToSVGFlags(s string) (SVGFlags, error) {
	for i := 0; i < len(_SVGFlags_index)-1; i++ {
		if s == _SVGFlags_name[_SVGFlags_index[i]:_SVGFlags_index[i+1]] {
			return SVGFlags(i + 35), nil
		}
	}
	return 0, errors.New("String: " + s + " is not a valid option for type: SVGFlags")
}
