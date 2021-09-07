package support

func True() *bool {
	value := true
	return &value
}

func False() *bool {
	return new(bool)
}
