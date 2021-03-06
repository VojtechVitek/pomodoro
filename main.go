package main

import (
	"bufio"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gosuri/uiprogress"
)

const (
	pomodoroTime = 25 * time.Second
	breakTime    = 5 * time.Second
)

// TODO: progress bar per pomodoro
// TODO: msg per pomodoro
// TODO: remove break pomodoros
type Pomodoro struct {
	bar     *uiprogress.Bar
	running bool
	done    chan struct{}
}

func NewPomodoro() *Pomodoro {
	pomodoro := Pomodoro{
		done: make(chan struct{}, 0),
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for {
			select {
			case <-pomodoro.done:
			case <-sig:
				if pomodoro.running {
					pomodoro.Stop()
				} else {
					os.Exit(0)
				}
			}
		}
	}()

	uiprogress.Start()

	return &pomodoro
}

func (p *Pomodoro) Run(duration int) bool {
	tick := time.NewTicker(time.Second / 10)
	p.running = true
	p.bar = uiprogress.AddBar(10 * duration).AppendCompleted().PrependElapsed()
	/*.PrependFunc(func(b *uiprogress.Bar) string {
		return p.msg + ": "
	})*/

	defer func() {
		tick.Stop()
		p.running = false
		p.done = make(chan struct{}, 0)
	}()

	for {
		select {
		case <-tick.C:
			p.bar.Set(p.bar.Current() + 1)
			if p.bar.Current() >= 10*duration {
				p.running = false
				close(p.done)
				return true
			}
		case <-p.done:
			return false
		}
	}
}

func (p *Pomodoro) Stop() bool {
	if !p.running {
		return false
	}
	close(p.done)
	return true
}

func main() {
	print("\033[H\033[2J")
	stdin := bufio.NewReader(os.Stdin)

	pomodoro := NewPomodoro()
	for {
		// Pomodoro.
		if !pomodoro.Run(25 * 60) {
			print("    Hit enter to start pomodoro..")
			stdin.ReadString('\n')
			print("\033[H\033[2J")
			continue
		}
		print("\a    Hit enter to start break..")
		stdin.ReadString('\n')
		print("\033[H\033[2J")

		// Break.
		pomodoro.Run(5 * 60)
		print("\a    Hit enter to start pomodoro..")
		stdin.ReadString('\n')
		print("\033[H\033[2J")
	}
}
