package proxy

import "os"

var LOG *Logger

func init() {
	LOG = NewLogger(os.Stdout)
}
