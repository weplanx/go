package basic

func True() *bool {
	value := true
	return &value
}

func False() *bool {
	value := false
	return &value
}
