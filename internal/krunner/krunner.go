package krunner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type ProgramService interface {
	SendNotification(string, string) error
}

type GoService struct{}

type MyService struct {
	programService ProgramService
}

//Method that can be mocked
func (gop GoService) SendNotification(user, content string) error {
	fmt.Printf("Hey %v this is your new information: %v\n", user, content)
	return nil
}

//Method that can be mocked
func (a MyService) StoreInformation(dbName, content string) error {
	a.programService.SendNotification("goremoteuser", "twenty encrypted files")
	fmt.Println("Storing changes and sending noification")
	return nil
}

func execute(sigch chan os.Signal, prog string, args ...string) error {
	fmt.Fprintf(os.Stderr, "execute: %#v %#v\n", prog, args)
	//context that is canceled when cancel() is called
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, prog, args...)
	//to kill a process and all its childs
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	defer syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) // negative pid = sending a signal to a Process Group

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case s := <-sigch:
			switch s {
			case syscall.SIGCHLD:
				return nil
			case syscall.SIGINT, syscall.SIGTERM:
				cancel()
			}
		}
	}
}

func Run(args []string) {
	fmt.Println(args)
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGCHLD, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigch)

	for {
		if err := execute(sigch, args[0], args[1:]...); err != nil {
			log.Fatal(err)
			os.Exit(0)
		}
	}
}
