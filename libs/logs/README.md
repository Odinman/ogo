## logs
logs is a Go logs manager. It can use many logs adapters. The repo is inspired by `database/sql` .


## What adapters are supported?

As of now this logs support console, file,smtp and conn.


## How to use it?

First you must import it

	import (
		"./libs/logs"
	)

Then init a Log (example with console adapter)

	log := NewLogger(10000)
	log.SetLogger("console", "")	

> the first params stand for how many channel

Use it like this:	
	
	log.Trace("trace")
	log.Info("info")
	log.Warn("warning")
	log.Debug("debug")
	log.Critical("critical")


## File adapter

Configure file adapter like this:

	log := NewLogger(10000)
	log.SetLogger("file", `{"filename":"test.log"}`)


## Conn adapter

Configure like this:

	log := NewLogger(1000)
	log.SetLogger("conn", `{"net":"tcp","addr":":7020"}`)
	log.Info("info")


## Smtp adapter

Configure like this:

	log := NewLogger(10000)
	log.SetLogger("smtp", `{"username":"odinmanlee@gmail.com","password":"xxxxxxxx","host":"smtp.gmail.com:587","sendTos":["xxx@gmail.com"]}`)
	log.Critical("sendmail critical")
	time.Sleep(time.Second * 30)
