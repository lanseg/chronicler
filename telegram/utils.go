package telegram

import (
	tgbot "github.com/lanseg/tgbot"
)

func GetLargestImage(sizes []*tgbot.PhotoSize) *tgbot.PhotoSize {
	if len(sizes) == 0 {
		return nil
	}

	var result *tgbot.PhotoSize = sizes[0]
	resultSize := int64(0)
	for _, photo := range sizes {
		size := photo.Width * photo.Height
		if size > resultSize {
			result = photo
			resultSize = size
		}
	}
	return result
}
