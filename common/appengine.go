package common

import (	
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"strings"
)

func Version(c context.Context) string {	
	version:=appengine.VersionID(c)
	array:=strings.Split(version,".")
	VERSION = array[0]
	return VERSION
}