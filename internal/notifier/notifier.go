package notifier

import (
	"io"
	"log"
	"os/exec"
)

type Notifier struct {
	fn func(*log.Logger, string)
}

func (n *Notifier) Notify(l *log.Logger, msg string) {
	n.fn(l, msg)
}

// Default notifier sends desktop notifications
func Default() *Notifier {
	n := Notifier{}
	n.fn = func(l *log.Logger, msg string) {
		cmd := exec.Command("notify-send", "Gobinet error", msg)
		if output, err := cmd.CombinedOutput(); err != nil {
			l.Printf("error sending notification:\n%s", output)
		}
	}
	return &n
}

func Writer(w io.Writer) *Notifier {
	n := Notifier{}
	n.fn = func(l *log.Logger, msg string) {
		if _, err := w.Write([]byte(msg)); err != nil {
			l.Printf("error sending notification:\n%s", err.Error())
		}
	}
	return &n
}
