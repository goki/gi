// Code generated by "stringer -type=TabViewSignals"; DO NOT EDIT.

package gi

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TabSelected-0]
	_ = x[TabAdded-1]
	_ = x[TabDeleted-2]
	_ = x[TabViewSignalsN-3]
}

const _TabViewSignals_name = "TabSelectedTabAddedTabDeletedTabViewSignalsN"

var _TabViewSignals_index = [...]uint8{0, 11, 19, 29, 44}

func (i TabViewSignals) String() string {
	if i < 0 || i >= TabViewSignals(len(_TabViewSignals_index)-1) {
		return "TabViewSignals(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TabViewSignals_name[_TabViewSignals_index[i]:_TabViewSignals_index[i+1]]
}

func (i *TabViewSignals) FromString(s string) error {
	for j := 0; j < len(_TabViewSignals_index)-1; j++ {
		if s == _TabViewSignals_name[_TabViewSignals_index[j]:_TabViewSignals_index[j+1]] {
			*i = TabViewSignals(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: TabViewSignals")
}