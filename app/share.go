package app

func Share(pattern string, recipient string) error {
	return shareService.ShareByEmail(pattern, recipient)
}
