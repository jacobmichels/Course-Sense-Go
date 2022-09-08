package notifier

import (
	"context"
	"fmt"
	"log"
	"net/smtp"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
)

var _ coursesense.Notifier = Email{}

type Email struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmail(host, username, password, from string, port int) Email {
	return Email{host, port, username, password, from}
}

func (e Email) Notify(ctx context.Context, section coursesense.Section, watchers ...coursesense.Watcher) error {
	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	msg := []byte(fmt.Sprintf("Hello from Course Sense!\n\nSpace has been found in the following course section: %s %d %s %s. Get over to WebAdvisor to claim the spot!", section.Course.Department, section.Course.Code, section.Code, section.Term))

	for _, watcher := range watchers {
		if watcher.Email == "" {
			continue
		}

		err := smtp.SendMail(fmt.Sprintf("%s:%d", e.host, e.port), auth, e.from, []string{watcher.Email}, msg)
		if err != nil {
			return fmt.Errorf("failed to notify %s: %w", watcher.Email, err)
		}
		log.Println("Notification email sent")
	}

	return nil
}
