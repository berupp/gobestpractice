package packagelog_test

import (
	"log"
	"minimalgo/packagelog"
	"testing"
)

func TestMyCoolFunction(t *testing.T) {
	packagelog.MyCoolFunction("Paul", 43) //Logs nothing
	packagelog.SetLogger(log.Default())
	packagelog.MyCoolFunction("Jill", 84) //Logs using standard library logger
}
