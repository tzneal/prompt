// Code generated by "stringer -type=itemType"; DO NOT EDIT

package prompt

import "fmt"

const _itemType_name = "itemChanCloseitemErroritemWorditemSemiitemQuotedStringitemLineContitemFilenameitemPlaceholderitemCompletionTypeitemPipeitemRAngleitemEOF"

var _itemType_index = [...]uint8{0, 13, 22, 30, 38, 54, 66, 78, 93, 111, 119, 129, 136}

func (i itemType) String() string {
	if i >= itemType(len(_itemType_index)-1) {
		return fmt.Sprintf("itemType(%d)", i)
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}
