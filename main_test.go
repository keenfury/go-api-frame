package main

import (
	"reflect"
	"testing"
)

func TestName_NameConverter(t *testing.T) {
	tests := []struct {
		name           string
		nameStruct     Name
		wantNameStruct Name
	}{
		{
			"CamelCase",
			Name{
				RawName: "User",
			},
			Name{
				RawName:    "User",
				Camel:      "User",
				LowerCamel: "user",
				Lower:      "user",
				AllLower:   "user",
				Abbr:       "use",
				Upper:      "USER",
			},
		},
		{
			"SnakeCase",
			Name{
				RawName: "api_user",
			},
			Name{
				RawName:    "api_user",
				Camel:      "ApiUser",
				LowerCamel: "apiUser",
				Lower:      "api_user",
				AllLower:   "apiuser",
				Abbr:       "api",
				Upper:      "APIUSER",
			},
		},
		{
			"SnakeCasePlus",
			Name{
				RawName: "_api_user_",
			},
			Name{
				RawName:    "_api_user_",
				Camel:      "ApiUser",
				LowerCamel: "apiUser",
				Lower:      "_api_user_",
				AllLower:   "apiuser",
				Abbr:       "_ap",
				Upper:      "APIUSER",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nameStruct.NameConverter()
			if !reflect.DeepEqual(tt.nameStruct, tt.wantNameStruct) {
				t.Errorf("NameConverter - %s: wanted: %+v; got: %+v", tt.name, tt.wantNameStruct, tt.nameStruct)
			}
		})
	}
}
