package main

import (
	"cert/aliyun"
	"cert/dnspod"
)

func main() {

	aliyun.Aliyun_cert("account", "domain")
	dnspod.Dnspod_cert("account", "domain")
}
