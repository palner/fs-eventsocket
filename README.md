# eventsocket

**Forked from [fiorix/go-eventsocket](https://github.com/fiorix/go-eventsocket)**

FreeSWITCH [Event Socket](http://wiki.freeswitch.org/wiki/Event_Socket) library
for the [Go programming language](http://golang.org).

It supports both inbound and outbound event socket connections, acting either
as a client connecting to FreeSWITCH or as a server accepting connections
from FreeSWITCH to control calls.

This code has not been tested in production and is considered alpha. Use at
your own risk.

## Installing

Make sure $GOPATH is set, and use the following command to install:

	go get github.com/palner/fs-eventsocket/eventsocket

The library is currently a single file, so feel free to drop into any project
without bothering to install.

## Main Differences

Main differences between this fork and [fiorix/go-eventsocket](https://github.com/fiorix/go-eventsocket):

- LogPrettyPrint()
- SendEvent()

### LogPrettyPrint

LogPrettyPrint prints Event headers and body to the logger; assuming of course you have a log set-up.

Example:

```go
	fs.Send("events json ALL")
	for {
		ev, err := fs.ReadEvent()
		if err != nil {
			log.Fatal(err)
			break
		}

		switch ev.Get("Event-Name") {
		case "HEARTBEAT":
			go HandleHeartbeat(ev)
		case "CHANNEL_HANGUP", "CUSTOM", "CHANNEL_ANSWER", "RELOADXML", "RECORD_STOP":
			log.Println("New event:", ev.Get("Event-Name"))
			ev.LogPrettyPrint()
		default:
			log.Println("New event:", ev.Get("Event-Name"))
		}


	}

	fs.Close()
```

### SendEvent

Sends FreeSWITCH a `sendevent` command, perfect for sending a NOTIFY reboot, etc.

Example:

```go
	fs, err := eventsocket.Dial([FS ADDRESS], [FS PASSWORD])
	if err != nil {
		log.Fatal("Whomp. FreesSWITCH connect error:", err.Error())
	}

	\\ get a uuid
	uuid, err := fs.Send("api create_uuid")
	if err != nil {
		fs.Close()
		log.Fatal("uuid errod", err.Error())
	}

	var eventMSG = eventsocket.MSG{
		"profile":      "internal",
		"event-string": "check-sync;reboot=true",
		"user":         "1234",
		"host":         "[DOMAIN of FreeSWITCH]",
		"call-id":      uuid.Body,
		"uuid":         uuid.Body,
		"to-uri":       "sip:1234@[IP of Endpoint or Proxy]",
		"from-uri":     "sip:1234@[DOMAIN of FreeSWITCH]",
	}

	response, err := fs.SendEvent("NOTIFY", eventMSG, "")
	if err != nil {
		log.Println("error sending notify:", err.Error())
	} else {
		log.Println("response:")
		response.LogPrettyPrint()
	}

	fs.Close()
```

## Usage

There are simple and clear examples of usage under the *examples* directory. A
client that connects to FreeSWITCH and originate a call, pointing to an
Event Socket server, which answers the call and instructs FreeSWITCH to play
an audio file.
