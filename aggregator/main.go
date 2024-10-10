package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	pbUser "e-wallet/proto/user"
	pbWallet "e-wallet/proto/wallet"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Aggregator struct {
	userClient   pbUser.UserServiceClient
	walletClient pbWallet.WalletServiceClient
}

func NewAggregator() *Aggregator {
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	walletConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return &Aggregator{
		userClient:   pbUser.NewUserServiceClient(userConn),
		walletClient: pbWallet.NewWalletServiceClient(walletConn),
	}
}

func (a *Aggregator) CreateUser(c *gin.Context) {
	var req pbUser.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := a.userClient.CreateUser(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (a *Aggregator) GetUser(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userId, _ := strconv.Atoi(userIdStr)
	res, err := a.userClient.GetUser(context.Background(), &pbUser.GetUserRequest{UserId: int32(userId)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (a *Aggregator) TopUp(c *gin.Context) {
	var req pbWallet.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := a.walletClient.TopUp(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (a *Aggregator) Transfer(c *gin.Context) {
	var req pbWallet.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := a.walletClient.Transfer(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (a *Aggregator) GetTransactions(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userId, _ := strconv.Atoi(userIdStr)
	res, err := a.walletClient.GetTransactions(context.Background(), &pbWallet.GetTransactionsRequest{UserId: int32(userId)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func main() {
	aggregator := NewAggregator()

	r := gin.Default()
	r.POST("/users", aggregator.CreateUser)
	r.GET("/users/:user_id", aggregator.GetUser)
	r.POST("/topup", aggregator.TopUp)
	r.POST("/transfer", aggregator.Transfer)
	r.GET("/transactions/:user_id", aggregator.GetTransactions)

	r.Run(":8080")
}
