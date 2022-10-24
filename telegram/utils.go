package telegram

func GetLargestImage(sizes []*PhotoSize) *PhotoSize {
	if len(sizes) == 0 {
		return nil
	}

	var result *PhotoSize = sizes[0]
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
