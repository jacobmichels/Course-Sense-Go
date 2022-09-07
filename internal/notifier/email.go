package notifier

import (
	"context"
	"fmt"
	"log"
	"net/smtp"

	"github.com/jacobmichels/Course-Sense-Go/internal/types"
)

type EmailNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailNotifier(host, username, password, from string, port int) *EmailNotifier {
	return &EmailNotifier{host, port, username, password, from}
}

func (en *EmailNotifier) Name() string {
	return "EmailNotifier"
}

func (en *EmailNotifier) Notify(ctx context.Context, section *types.CourseSection) error {
	auth := smtp.PlainAuth("", en.username, en.password, en.host)
	msg := []byte(fmt.Sprintf("Hello from Course Sense!\n\nSpace has been found in the following course section: %s %d %s %s. Get over to WebAdvisor to claim the spot!", section.Department, section.CourseCode, section.SectionCode, section.Term))

	for _, watcher := range section.Watchers {
		err := smtp.SendMail(fmt.Sprintf("%s:%d", en.host, en.port), auth, en.from, []string{watcher.Email}, msg)
		if err != nil {
			return fmt.Errorf("failed to notify %s: %w", watcher.Email, err)
		}
		log.Println("mail sent")
	}

	return nil
}
