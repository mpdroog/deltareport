Deltareport
=============
Find file/dir changes and queue to Beanstalkd for processing.

> This application does not spawn off go-routines, it does all in the main-routine.

> Please don't run this application multiple times with the same delta.db!

config.json
```
{
	"Files": {
		"./test.txt": {
			"To": "admin",
			"Recurse": false
		},
		"./test.d": {
			"To": "sess",
			"Recurse": true,
			"IncludeExt": [
				".txt", ".log"
			]
		}
	},
	"Queues": {
		"mail": {
			"admin": {
				"Beanstalkd": "127.0.0.1:11300",
				"From": "support",
				"To": ["errors@itshosted.nl"],
				"Subject": "[AUTOGEN] "
			}
		},
		"newline": {
			"sess": {
				"Beanstalkd": "127.0.0.1:11300",
				"Queue": "sess"
			}
		}
	}
}
```
This example config scans for changes:

* Diff the textfile `./test.txt` and e-mails diff using SMTPw (https://github.com/mpdroog/smtpw).
* Diff all files in `./test.d` and write (messages separated by newline) changes to to sess-queue

How?
=============
Using the keyvaluestore (Bolt) to remember the last read position
and on change read all changes and write these to the assigned queue.
It reads/loads it's status from `./delta.db`.

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
	Path string
	Line string
}
```