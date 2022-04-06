package common

type Messenger interface {
	SendMessage(URL, message, title, content string) (error, []byte)
	SendPhoto(URL, message, fileName, title string, photo []byte) (error, []byte)
}
