// Code generated by "stringer -type=MenuButtonTypes"; DO NOT EDIT.

package gi

import (
	"errors"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MenuButtonFilled-0]
	_ = x[MenuButtonOutlined-1]
	_ = x[MenuButtonText-2]
	_ = x[MenuButtonTypesN-3]
}

const _MenuButtonTypes_name = "MenuButtonFilledMenuButtonOutlinedMenuButtonTextMenuButtonTypesN"

var _MenuButtonTypes_index = [...]uint8{0, 16, 34, 48, 64}

func (i MenuButtonTypes) String() string {
	if i < 0 || i >= MenuButtonTypes(len(_MenuButtonTypes_index)-1) {
		return "MenuButtonTypes(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MenuButtonTypes_name[_MenuButtonTypes_index[i]:_MenuButtonTypes_index[i+1]]
}

func (i *MenuButtonTypes) FromString(s string) error {
	for j := 0; j < len(_MenuButtonTypes_index)-1; j++ {
		if s == _MenuButtonTypes_name[_MenuButtonTypes_index[j]:_MenuButtonTypes_index[j+1]] {
			*i = MenuButtonTypes(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: MenuButtonTypes")
}

var _MenuButtonTypes_descMap = map[MenuButtonTypes]string{
	0: `MenuButtonFilled represents a filled MenuButton with a background color and no border`,
	1: `MenuButtonOutlined represents an outlined MenuButton with a border on all sides and no background color`,
	2: `MenuButtonText represents a MenuButton with no border or background color.`,
	3: ``,
}

func (i MenuButtonTypes) Desc() string {
	if str, ok := _MenuButtonTypes_descMap[i]; ok {
		return str
	}
	return "MenuButtonTypes(" + strconv.FormatInt(int64(i), 10) + ")"
}