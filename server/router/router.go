package router

import (
	"server/controller"

	"github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func Router() *gin.Engine {
	//1.创建路由
	router := gin.Default()
    // 创建一个自定义的CORS配置  
    config := cors.DefaultConfig()  
    config.AllowOrigins = []string{"http://localhost:9528"} // 允许的前端应用URL  
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}  
    config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}  
    config.AllowCredentials = true // 允许携带凭证  
   
     // 使用自定义配置初始化CORS中间件并应用到所有路由  
     router.Use(cors.New(config)) 
	/*
	 * Group
	 */
	//2.绑定路由规则，执行的函数
	user := router.Group("user")
	{
		user.GET("/login", controller.UserController{}.Login)
		user.POST("/setwallet", controller.UserController{}.SetWallet)
		user.POST("/getwallet", controller.UserController{}.GetWallet)
		user.POST("/submitOrder", controller.UserController{}.SubmitProposal)
		user.POST("/updateOrder", controller.UserController{}.UpdateOrder)
	}
	good := router.Group("good")
	{
		good.GET("/getAll", controller.GoodController{}.GetAllGoods)
		good.POST("/getGood", controller.GoodController{}.GetGood)
		good.POST("/getGoodByOwner", controller.GoodController{}.GetGoodByOwner)
		good.POST("/updateGoodPrice", controller.GoodController{}.UpdateGoodPrice)
	}
	proposal := router.Group("proposal")
	{
		proposal.POST("/getProposal", controller.ProposalController{}.GetProposal)
		proposal.POST("/getProposalBySeller", controller.ProposalController{}.GetProposalBySeller)
		proposal.POST("/getProposalByBuyer", controller.ProposalController{}.GetProposalByBuyer)
		proposal.POST("/getProposalByOrderNum", controller.ProposalController{}.GetProposalByOrderNum)
		proposal.POST("/setProposal", controller.ProposalController{}.SetProposal)
		proposal.POST("/updateProposal", controller.ProposalController{}.UpdateProposal)
	}
	return router
}
