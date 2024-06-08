package main

import(
	"server/router"
)

func main(){
	r:=router.Router()
    // 监听端口，默认8080
    r.Run(":8080")
}