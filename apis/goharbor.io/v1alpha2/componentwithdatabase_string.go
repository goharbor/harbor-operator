// Code generated by "stringer -type=ComponentWithDatabase -linecomment"; DO NOT EDIT.

package v1alpha2

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CoreDatabase-0]
	_ = x[NotaryServerDatabase-1]
	_ = x[NotarySignerDatabase-2]
	_ = x[ClairDatabase-3]
}

const _ComponentWithDatabase_name = "corenotaryservernotarysignerclair"

var _ComponentWithDatabase_index = [...]uint8{0, 4, 16, 28, 33}

func (i ComponentWithDatabase) String() string {
	if i < 0 || i >= ComponentWithDatabase(len(_ComponentWithDatabase_index)-1) {
		return "ComponentWithDatabase(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ComponentWithDatabase_name[_ComponentWithDatabase_index[i]:_ComponentWithDatabase_index[i+1]]
}
