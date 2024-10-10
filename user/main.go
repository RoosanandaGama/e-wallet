package main

import (
	"context"
	"log"
	"net"

	pb "e-wallet/proto/user"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedUserServiceServer
	DB *gorm.DB
}

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Email string `gorm:"size:100;unique"`
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := User{Name: req.GetName(), Email: req.GetEmail()}
	s.DB.Create(&user)
	return &pb.CreateUserResponse{UserId: int32(user.ID), Message: "User created successfully"}, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var user User
	if err := s.DB.First(&user, req.GetUserId()).Error; err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{UserId: int32(user.ID), Name: user.Name, Email: user.Email}, nil
}

func main() {
	dsn := "host=localhost user=ewallet_service password=ewallet_password dbname=ewallet_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database")
	}

	db.AutoMigrate(&User{})

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &server{DB: db})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen:", err)
	}
	log.Println("User Service running on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}
}
