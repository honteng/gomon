package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/honteng/gomon/notify"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/c9s/gomon/logger"
	"github.com/howeyc/fsnotify"
)

var (
	versionStr = "1.0.0"

	wasFailed                    = true
	notifier     notify.Notifier = nil
	alwaysNotify                 = false
)

type FileBasedTaskRunner struct {
	cmd            Command
	AppendFilename bool
	Chdir          bool
	filenameCh     chan string
}

func notifyError(d time.Duration, err error) {
	if err != nil {
		logger.Errorln("Task Failed:", err.Error())

		notifier.NotifyFailed("Build Failed", err.Error())
	} else {
		logger.Infoln("Task Completed:", d)

		if wasFailed {
			wasFailed = false
			notifier.NotifyFixed("Build Fixed", fmt.Sprintf("Spent: %s", d))
		} else if alwaysNotify {
			notifier.NotifySucceeded("Build Succeeded", fmt.Sprintf("Spent: %s", d))
		}
	}
}

func (r *FileBasedTaskRunner) loop() {
	var runner *CommandRunner
	var g *errgroup.Group
	for {
		select {
		case filename := <-r.filenameCh:
			var chdir = ""
			if r.Chdir {
				chdir = filepath.Dir(filename)
			}

			var args []string
			if r.AppendFilename {
				args = append(args, filename)
			}

			logger.Infof("Starting: chdir=%s commands=%v args=%v", chdir, r.cmd, args)
			if runner != nil {
				runner.Stop()
				g.Wait()
			}
			logger.Debug("wait done")

			runner = &CommandRunner{}
			ch := make(chan struct{})
			g, _ = errgroup.WithContext(context.Background())
			f := func(args []string, chdir string) func() error {
				return func() error {
					t := time.Now()
					logger.Debug("runner starts")
					defer logger.Debug("runner ends")
					runner.Start(r.cmd, args, chdir)
					ch <- struct{}{}
					err := runner.Wait(r.cmd, args, chdir)
					if err != nil && strings.HasPrefix(err.Error(), "exit status ") {
						logger.Error("exit")
						notifyError(time.Since(t), err)
					}
					return err
				}
			}(args, chdir)

			g.Go(f)

			<-ch
		}
	}
}

func (r *FileBasedTaskRunner) Run(filename string) (duration time.Duration, err error) {
	r.filenameCh <- filename
	return 0, nil
}

func main() {
	dirArgs, cmdArgs := options.Parse(os.Args)
	dirArgs = FilterExistPaths(dirArgs)

	var matchAll = false

	if options.Bool("h") {
		fmt.Println("Usage: gomon [options] [dir] [-- command]")
		for _, option := range options {
			if _, ok := option.value.(string); ok {
				fmt.Printf("  -%s=%s: %s\n", option.flag, option.value, option.description)
			} else {
				fmt.Printf("  -%s: %s\n", option.flag, option.description)
			}
		}
		os.Exit(0)
	}
	if options.Bool("v") {
		fmt.Printf("gomon %s\n", versionStr)
		os.Exit(0)
	}

	if options.Bool("install-growl-icons") {
		notify.InstallGrowlIcons()
		os.Exit(0)
		return
	}

	matchAll = options.Bool("matchall")
	alwaysNotify = options.Bool("alwaysnotify")

	if options.Bool("d") {
		logger.Instance().SetLevel(logrus.DebugLevel)
	}

	cmd := Command(cmdArgs)

	if len(dirArgs) == 0 {
		var cwd, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dirArgs = []string{cwd}
	}

	if runtime.GOOS == "darwin" {
		logger.Infoln("Setting up Notification Center for OS X ...")
		notifier = notify.NewOSXNotifier()
	}
	if notifier == nil {
		if _, err := os.Stat("/Applications/Growl.app"); err == nil {
			logger.Infoln("Found Growl.app, setting up GNTP notifier...")
			notifier = notify.NewGNTPNotifier(options.String("gntp"), "gomon")
		}
	}
	if notifier == nil {
		notifier = notify.NewTextNotifier()
	}

	logger.Infoln("Watching", dirArgs, "for", cmd)

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	for _, dir := range dirArgs {
		if options.Bool("R") {
			subfolders := Subfolders(dir)
			for _, f := range subfolders {
				err = watcher.WatchFlags(f, fsnotify.FSN_ALL)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			err = watcher.WatchFlags(dir, fsnotify.FSN_ALL)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	var taskRunner = &FileBasedTaskRunner{
		cmd:            cmd,
		AppendFilename: options.Bool("F"),
		Chdir:          options.Bool("chdir"),
		filenameCh:     make(chan string),
	}

	go taskRunner.loop()

	var patternStr string = options.String("m")
	if len(patternStr) == 0 {
		// the empty regexp matches everything anyway
		matchAll = true
	}

	var pattern = regexp.MustCompile(patternStr)

	for {
		select {
		case e := <-watcher.Event:
			var matched = matchAll
			if !matched {
				matched = pattern.MatchString(e.Name)
			}

			if !matched {
				if options.Bool("d") {
					logger.Debugf("Ignored file=%s", e)
				}
				continue
			}

			if options.Bool("d") {
				logger.Debugf("Event=%+v", e)
			} else {
				if e.IsCreate() {
					logger.Infoln("Created", e.Name)
				} else if e.IsModify() {
					logger.Infoln("Modified", e.Name)
				} else if e.IsDelete() {
					logger.Infoln("Deleted", e.Name)
				} else if e.IsRename() {
					logger.Infoln("Renamed", e.Name)
				}
			}

			filename := e.Name
			taskRunner.Run(filename)

		case err := <-watcher.Error:
			log.Println("Error:", err)
		}
	}

	watcher.Close()
}
