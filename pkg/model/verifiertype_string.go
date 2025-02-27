// Code generated by "stringer -type=VerifierType --trimprefix=Verifier"; DO NOT EDIT.

package model

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[verifierUnknown-0]
	_ = x[VerifierNoop-1]
	_ = x[VerifierDeterministic-2]
	_ = x[verifierDone-3]
}

const _VerifierType_name = "verifierUnknownNoopDeterministicverifierDone"

var _VerifierType_index = [...]uint8{0, 15, 19, 32, 44}

func (i VerifierType) String() string {
	if i < 0 || i >= VerifierType(len(_VerifierType_index)-1) {
		return "VerifierType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _VerifierType_name[_VerifierType_index[i]:_VerifierType_index[i+1]]
}
