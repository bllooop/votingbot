package logging

import (
	"os"

	"github.com/rs/zerolog"
)

var Log = zerolog.New(os.Stdout).Level(zerolog.TraceLevel)
