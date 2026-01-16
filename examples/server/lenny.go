/**

Lenny via ESL
Copyright (C) 2026 Fred Posner. All Rights Reserved.
Copyright (C) 2026 The Palner Group, Inc. All Rights Reserved.
License: MIT

Building:

GOOS=linux GOARCH=amd64 go build -o go-lenny
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o go-lenny
GOOS=linux GOARCH=arm GOARM=7 go build -o go-lenny-pi

This is an example for a go "lenny" using freeswitch ESL. In your freeswitch
dialplan, you would reach Lenny via something like...

	<extension name="testing_playground">
		<condition field="destination_number" expression="^lenny$">
			<action application="socket" data="127.0.0.1:9090"/>
			<action application="hangup"/>
		</condition>
	</extension>

In this example, the "hello" greeting is appFolder+"Am-hello.wav" and the random
soundfiles to play are in a slice in the GetAmSound function.

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"

	"github.com/palner/fs-eventsocket/eventsocket"
)

var (
	appFolder   string
	logFile     string
	logFileLine bool
)

func init() {
	flag.StringVar(&logFile, "log", "/var/log/go-lenny.log", "location of log file or - for stdout")
	flag.BoolVar(&logFileLine, "logextra", false, "add filename to log")
	flag.StringVar(&appFolder, "folder", "/usr/share/freeswitch/sounds/Am/", "default location of app/config")
}

func main() {
	defer os.Exit(0)

	// get flags
	flag.Parse()

	// init log
	if logFile != "-" && logFile != "stdout" {
		lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Panic(err)
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			runtime.Goexit()
		}

		defer lf.Close()
		if logFileLine {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
		} else {
			log.SetFlags(log.LstdFlags)
		}

		log.SetOutput(lf)
		log.Println("== started ==")
	}

	// esl listenger
	eventsocket.ListenAndServe(":9090", handler)
}

func handler(c *eventsocket.Connection) {
	log.Println("[+] new call:", c.RemoteAddr())
	c.Send("connect")
	c.Send("myevents")
	c.Execute("answer", "", false)
	_, err := c.Execute("playback", appFolder+"Am-hello.wav", false)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("[-] said hello")
	}

	timeoutcount := 0
	for {
		ev, err := c.ReadEvent()
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("[x] call most likely disconnected")
				break
			} else {
				log.Println("ERR", err.Error())
			}
		} else {
			log.Println("[-]", ev.Get("Unique-Id"), "New event", ev.Get("Event-Name"))
			if ev.Get("Event-Name") == "CHANNEL_EXECUTE_COMPLETE" {
				if ev.Get("Application") == "wait_for_silence" {
					skipfile := false
					log.Println("[-] wait for silence ended")
					if ev.Get("Variable_wait_for_silence_timeout") == "true" {
						log.Println("[-] ended with a timeout")
						if ev.Get("Variable_wait_for_silence_listenhits") != "0" {
							log.Println("[-] sound detected. send anotherwait for silence.", ev.Get("Variable_wait_for_silence_listenhits"))
							log.Println("[-] reset timeoutcount. was", timeoutcount)
							timeoutcount = 0
							_, err = c.Execute("wait_for_silence", "300 45 5 5000", false)
							if err != nil {
								log.Println("ERR waiting for silence:", err.Error())
								break
							}

							skipfile = true
						} else {
							timeoutcount++
							if timeoutcount > 1 {
								break
							}
						}
					}

					if !skipfile {
						file := GetAmSound()
						log.Println("[-] playing", file)
						_, err = c.Execute("playback", appFolder+file, true)
						if err != nil {
							log.Println("ERR playback:", err.Error())
							break
						}
					}
				} else if ev.Get("Application") == "playback" {
					if ev.Get("Application-Response") == "FILE PLAYED" {
						log.Println("[-] file played. send wait for silence.")
						_, err = c.Execute("wait_for_silence", "300 45 5 5000", false)
						if err != nil {
							log.Println("ERR waiting for silence:", err.Error())
							break
						}
					}
				}
			}
		}
	}

	log.Println("[x] ending call")
	c.Send("hangup")
	c.Send("exit")

	// listen:
	//
	//	_, err = c.Execute("wait_for_silence", "300 45 5 15000", true)
	//	if err != nil {
	//		log.Println(err)
	//		c.Send("exit")
	//	} else {
	//		log.Println("[-] waiting for silence")
	//	}
	//
	//	file := GetAmSound()
	//	_, err = c.Execute("playback", appFolder+file, true)
	//	if err != nil {
	//		log.Println(err)
	//		c.Send("exit")
	//	} else {
	//		log.Println("[-] played", file)
	//	}
	//
	//	goto listen
}

func GetAmSound() string {
	sounds := []string{"Am-call-else-please", "Am-callsomeoneelse", "Am-fbbbt", "Am-funny", "Am-huh", "Am-playcomputer", "Am-sayagain", "Am-sayagain2"}
	randomIndex := rand.Intn(len(sounds))
	randomPick := sounds[randomIndex]
	return randomPick + ".wav"
}
