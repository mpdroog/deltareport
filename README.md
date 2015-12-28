Deltareport
=============
Find file/dir changes and queue to Beanstalkd for processing.

> This application does not spawn off go-routines, it does all in the main-routine.

> Please don't run this application multiple times with the same delta.db!

config.json
```
{
	"Files": [
		{
			"_": "Queue to mail/admin with everything that changed since last time",
			"Path": "./test.txt",                                                        // File to watch
			"To": "admin",                                                               // Queue on Queues.mail.admin
			"Recurse": false,                                                            // Path points to a file
			"Linediff": false                                                            // Queue all changes in 1 entry
		},
		{
			"_": "Queue to newline/sess about any changes in subfiles and separate by newline",
			"Path": "./test.d",                                                          // Dir to watch
			"To": "sess",                                                                // Queue on Queues.newline.sess
			"Recurse": true,                                                             // Path points to a directory
			"IncludeExt": [                                                              // Extensions to watch for change
				".txt",
				".log"
			],
			"Linediff": true                                                             // Queue by newline(\n)
		},
		{
			"_": "Queue to newline/slack and write to #channel",
			"Path": "./test.d",
			"To": "sess",
			"Tags": ["channel"],                                                         // Write to #channel
			"Recurse": true,
			"IncludeExt": [
				".txt",
				".log"
			],
			"Linediff": false
		}
	],
	"Queues": {
		"mail": {
			"admin": {
				"Beanstalkd": "127.0.0.1:11300",                                         // Hostname:port to beanstalkd
				"From": "support",
				"To": ["errors@itshosted.nl"],
				"Subject": "[AUTOGEN] "                                                  // Subject prefix
			}
		},
		"newline": {
			"sess": {
				"Beanstalkd": "127.0.0.1:11300",
				"Queue": "linediffs"                                                     // Beanstalkd tube
			},
			"slack": {
				"Beanstalkd": "127.0.0.1:11300",
				"Queue": "slack"
			}
		}
	},
	"Db": "/var/deltareport/example.db"                                                  // Database to save file pointers
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