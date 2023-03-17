auth-build:
	go build -o bin/auth github.com/sina-am/social-media/internal/auth 

chat-build:
	go build -o bin/chat github.com/sina-am/social-media/internal/chat

feed-build:
	go build -o bin/feed github.com/sina-am/social-media/internal/feed 

all: feed auth chat