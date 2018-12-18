/*
Copyright 2018 Ridecell, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"encoding/json"
)

// Intercept JSON decoding and try to deal with "simple" values before giving
// up and assuming it's a full struct. This allows things like:
//
//    config:
//      foo: bar
//      baz: false
//
// in a config section. This is all because the Kubernetes codegen machinery
// can't cope with a map[string]interface{}, since it could be some composite
// type, which would break all kinds of things.
func (v *ConfigValue) UnmarshalJSON(data []byte) error {
	var tmp interface{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		// Wat?
		return err
	}
	boolVal, ok := tmp.(bool)
	if ok {
		v.Bool = &boolVal
		return nil
	}
	intVal, ok := tmp.(int)
	if ok {
		v.Int = &intVal
		return nil
	}
	stringVal, ok := tmp.(string)
	if ok {
		v.String = &stringVal
		return nil
	}
	// It was something else.
	return json.Unmarshal(data, v)
}

// Run the reverse, convert the union back into an interface{} for use in JSON
// or YAML encoding when building the config file.
func (v *ConfigValue) ToNilInterface() interface{} {
	if v.Bool != nil {
		return *v.Bool
	} else if v.Int != nil {
		return *v.Int
	} else if v.String != nil {
		return *v.String
	} else {
		panic("Unknown ConfigValue type")
	}
}
