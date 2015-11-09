Deltareport
=============
Read all `Files` (config.json) and write changes since the last check to `STDOUT`.

config.json
```
{
	"Files": [
		"./test.txt"
	]
}
```

How?
=============
Using the keyvaluestore (Bolt) to remember the last read position
and on change read all changes and write these to `STDOUT`.
