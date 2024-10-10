package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "e-wallet/proto/wallet"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedWalletServiceServer
	DB *gorm.DB
}

type Wallet struct {
	ID      uint    `gorm:"primaryKey"`
	UserID  uint    `gorm:"index"`
	Balance float32 `gorm:"default:0"`
}

type Transaction struct {
	ID              uint `gorm:"primaryKey"`
	UserID          int32
	Amount          float32
	TransactionType string
	CreatedAt       time.Time
}

func (s *server) TopUp(ctx context.Context, req *pb.TopUpRequest) (*pb.TopUpResponse, error) {
	var wallet Wallet
	// Mencari wallet berdasarkan user_id
	result := s.DB.First(&wallet, "user_id = ?", req.GetUserId())

	if result.Error != nil {
		// Jika wallet tidak ditemukan, buat wallet baru
		if result.Error == gorm.ErrRecordNotFound {
			wallet = Wallet{
				UserID:  uint(req.GetUserId()),
				Balance: 0, // Saldo awal 0
			}
			// Simpan wallet baru ke database
			if err := s.DB.Create(&wallet).Error; err != nil {
				return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "failed to find wallet: %v", result.Error)
		}
	}

	// Update saldo wallet
	wallet.Balance += float32(req.GetAmount())
	if err := s.DB.Save(&wallet).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update wallet balance: %v", err)
	}

	// Simpan transaksi
	transaction := Transaction{
		UserID:          req.GetUserId(),
		Amount:          float32(req.GetAmount()),
		TransactionType: "topup",
		CreatedAt:       time.Now(),
	}
	if err := s.DB.Create(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	return &pb.TopUpResponse{Message: "Top-up successful"}, nil
}

func (s *server) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	var fromWallet, toWallet Wallet
	s.DB.First(&fromWallet, "user_id = ?", req.GetFromUserId())
	s.DB.First(&toWallet, "user_id = ?", req.GetToUserId())

	if fromWallet.Balance < float32(req.GetAmount()) {
		return &pb.TransferResponse{Message: "Insufficient funds"}, nil
	}

	fromWallet.Balance -= float32(req.GetAmount())
	toWallet.Balance += float32(req.GetAmount())
	s.DB.Save(&fromWallet)
	s.DB.Save(&toWallet)

	transaction := Transaction{UserID: req.GetFromUserId(), Amount: float32(req.GetAmount()), TransactionType: "transfer", CreatedAt: time.Now()}
	s.DB.Create(&transaction)

	return &pb.TransferResponse{Message: "Transfer successful"}, nil
}

func (s *server) GetTransactions(ctx context.Context, req *pb.GetTransactionsRequest) (*pb.GetTransactionsResponse, error) {
	var transactions []Transaction
	s.DB.Where("user_id = ?", req.GetUserId()).Find(&transactions)

	response := &pb.GetTransactionsResponse{}
	for _, t := range transactions {
		response.Transactions = append(response.Transactions, &pb.Transaction{
			Id:              int32(t.ID),
			Amount:          float32(t.Amount),
			TransactionType: t.TransactionType,
		})
	}

	return response, nil
}

func main() {
	dsn := "host=localhost user=ewallet_service password=ewallet_password dbname=ewallet_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database")
	}

	db.AutoMigrate(&Wallet{}, &Transaction{})

	grpcServer := grpc.NewServer()
	pb.RegisterWalletServiceServer(grpcServer, &server{DB: db})

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("failed to listen:", err)
	}
	log.Println("Wallet Service running on port 50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}
}
