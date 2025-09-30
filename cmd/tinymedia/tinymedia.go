package main

import (
	"flag"
	"fmt"
	"strings"
)

type listFlag []string

func (i *listFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *listFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var inputs listFlag

	var meta = flag.String("m", "", "-m=field1=value1,field2=value2")
	var metaVendor = flag.String("mv", "", "-mv=tinymeta")
	flag.Var(&inputs, "i", "input files")

	flag.Parse()

	metaFields := strings.Split(*meta, ",")
	if len(metaFields) > 1 || metaFields[0] != "" {
		handleMeta(inputs, metaFields, *metaVendor)
	}
}
