// Code generated by "stringer -type=completionType"; DO NOT EDIT

package prompt

import "fmt"

const _completionType_name = "completeNonecompletePartialcompleteExact"

var _completionType_index = [...]uint8{0, 12, 27, 40}

func (i completionType) String() string {
	if i >= completionType(len(_completionType_index)-1) {
		return fmt.Sprintf("completionType(%d)", i)
	}
	return _completionType_name[_completionType_index[i]:_completionType_index[i+1]]
}
