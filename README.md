Deltareport
=============
Find file/dir changes and queue to Beanstalkd for processing.

> This application does not spawn off go-routines, it does all in the main-routine.

> Please don't run this application multiple times with the same delta.db!

Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

config.toml
```
# Directory to scan for additional Files-entries
# allowing a more modular approach
Confdir = "./conf.d"
# Path to auto-created filed file for deltareport
# to remember it's state
Db = "/var/deltareport/example.db"

# Output queues
# The queue type 'mail' and 'newline' indicate
# what type of JSON is added to the Beanstalkd queue.
[Queues.mail]
	[Queues.mail.admin]
		Beanstalkd = "127.0.0.1:11300"
		From = "support"
		To = ["errors@itshosted.nl"]
		Subject = "[AUTOGEN] "

	[Queues.newline.sess]
		Beanstalkd = "127.0.0.1:11300"
		Queue = "linediffs"
	[Queues.newline.slack]
		Beanstalkd = "127.0.0.1:11300"
		Queue = "slack"

[[Files]]
	# Watch an individual file and send to queues.mail.admin any changed byte
	Path = "./test.txt"
	To = "admin"
	Recurse = false
	Linediff = false

[[Files]]
	# Watch a directory recursive and send changes files+lines to queues.newsline.sess
	# Filtering by file extension+regular expression

	Path = "./test.d"
	To = "sess"
	Recurse = true
	IncludeExt = [
		".txt",
		".log"
	]
	Regex = "/valid.txt$"
	# Add every changed line separately to the queue
	Linediff = true

[[Files]]
	# Queue to newline/slack and write to #channel
	# Filtering by file extension
	Path = "./test.d"
	To = "sess"
	Tags = ["channel"]
	Recurse = true
	IncludeExt = [
		".txt",
		".log"
	]
	# Add every changed file to the queue
	Linediff = false
```

Help
=============
- "Linediff": true
  If the diff is "msg1\nmsg2\nmsg3" then msg1/msg2 and msg3 are all added separately in the queue
  (creating 3 jobs in the queue instead of 1).
- "Tags": ["channel"]
  Metadata that only the worker understands.
- Watching same files/dirs multiple times?
  Yes possible.

How?
=============
Using the keyvaluestore (Bolt) to remember the last read position
and on change read all changes and write these to the assigned queue.
It reads/loads it's status from `./delta.db`.

Beanstalkd is used as a persistant queue between this application (diffing)
and processing (workers).

Workers?
=============
- SMTPw - https://github.com/mpdroog/smtpw
- Slackd - https://github.com/mpdroog/slackd

Datastructures
==============
```
type Email struct {
	From string                 // Key that MUST match From in config
	To []string
	Subject string
	Html string
	Text string
	HtmlEmbed map[string]string // file.png => base64(bytes)
}
```

```
type LineDiff struct {
	Hostname string
	Path     string
	Line     string
	Tags     []string
}
```