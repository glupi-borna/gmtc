package utils

func IsBetween(char byte, start byte, end byte) bool {
	return char >= start && char <= end
}

func IsWhitespace(char byte) bool {
	return char == ' ' || char == '\n' || char == '\t' || char == '\r'
}

func IsWhitespaceNoNL(char byte) bool {
	return char == ' ' || char == '\t'
}

func IsHexNumber(char byte) bool {
	return IsNumber(char) || IsBetween(char, 'a', 'f') || IsBetween(char, 'A', 'f')
}

func IsNumber(char byte) bool {
	return IsBetween(char, '0', '9')
}

func IsLetter(char byte) bool {
	return IsBetween(char, 'A', 'Z') || IsBetween(char, 'a', 'z')
}

func IsIdentChar(char byte, i int) bool {
	if IsLetter(char) || char == '_' {
		return true
	}
	return i != 0 && IsNumber(char)
}
