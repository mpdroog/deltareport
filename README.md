Deltareport
=============
Find file/dir changes and queue to Beanstalkd for processing.

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
			"Recurse": true
		}
	},
	"Queues": {
		"mail": {
			"admin": {
				"Beanstalkd": "127.0.0.1:11300",
				"From": "noreply@itshosted.nl",
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
* Diff all files in `./test.d` and write newline separated changes to to sess-queue

How?
=============
Using the keyvaluestore (Bolt) to remember the last read position
and on change read all changes and write these to the assigned queue.
