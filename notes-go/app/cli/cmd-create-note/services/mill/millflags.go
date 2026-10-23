package millflagscreatenote

import (
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/spf13/pflag"
)

type mill struct {
	plumprimitive.Mill
	flags *pflag.FlagSet
}

type Template struct {
	Flags *pflag.FlagSet
}

func New(params Template) *mill {
	return &mill{
		flags: params.Flags,
	}
}

// TODO make a resolver like state machine .. bref je me comprends
