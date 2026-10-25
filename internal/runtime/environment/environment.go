// Package environment Defines the runtime's environment structure
package environment

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/runtime"
)

var values = make(map[string]runtime.Value)

func Define(name string, value runtime.Value) {
	values[name] = value
}

func Access(name string) runtime.Value {
	if v, ok := values[name]; ok {
		return v
	}

	panic(fmt.Sprintf("Undefined variable %s.", name))
}
