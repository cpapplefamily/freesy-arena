// Code generated by "stringer -type input"; DO NOT EDIT.

package plc

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[fieldEstop-0]
	_ = x[redEstop1-1]
	_ = x[redEstop2-2]
	_ = x[redEstop3-3]
	_ = x[blueEstop1-4]
	_ = x[blueEstop2-5]
	_ = x[blueEstop3-6]
	_ = x[redConnected1-7]
	_ = x[redConnected2-8]
	_ = x[redConnected3-9]
	_ = x[blueConnected1-10]
	_ = x[blueConnected2-11]
	_ = x[blueConnected3-12]
	_ = x[inputCount-13]
}

const _input_name = "fieldEstopredEstop1redEstop2redEstop3blueEstop1blueEstop2blueEstop3redConnected1redConnected2redConnected3blueConnected1blueConnected2blueConnected3inputCount"

var _input_index = [...]uint8{0, 10, 19, 28, 37, 47, 57, 67, 80, 93, 106, 120, 134, 148}

func (i input) String() string {
	if i < 0 || i >= input(len(_input_index)-1) {
		return "input(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _input_name[_input_index[i]:_input_index[i+1]]
}
