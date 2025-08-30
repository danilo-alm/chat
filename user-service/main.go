package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"user-service/pb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type User struct {
	Id       string `bson:"_id"`
	Name     string `bson:"name"`
	Username string `bson:"username"`
}

type server struct {
	pb.UnimplementedUserServiceServer
	mongoClient *mongo.Client
}

const mongoDatabase = "chatdb"
const mongoCollection = "user"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	var client *mongo.Client

	mongo_uri := getEnv("MONGO_URI", "mongodb://root:example@mongo:27017")
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Trying to connect to MongoDB.")
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	log.Println("Successfully connected to MongoDB!")

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	userServer := &server{
		mongoClient: client,
	}

	pb.RegisterUserServiceServer(s, userServer)

	enable_reflection := getEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enable_reflection)
	if enable_reflection == "true" {
		reflection.Register(s)
	}

	log.Println("gRPC server started on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	collection := s.mongoClient.Database(mongoDatabase).Collection(mongoCollection)

	userInsert := User{
		Id:       req.User.GetId(),
		Name:     req.User.GetName(),
		Username: req.User.GetUsername(),
	}
	result, err := collection.InsertOne(ctx, userInsert)
	if err != nil {
		return nil, err
	}

	user := mapUserToPbUser(userInsert)
	user.Id = result.InsertedID.(primitive.ObjectID).Hex()

	return &pb.CreateUserResponse{Id: user.GetId()}, nil
}

func (s *server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserResponse, error) {
	return getUser(ctx, s.mongoClient, bson.M{"_id": req.GetId()})
}

func (s *server) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserResponse, error) {
	return getUser(ctx, s.mongoClient, bson.M{"username": req.GetUsername()})
}

func getUser(ctx context.Context, c *mongo.Client, filter any) (*pb.GetUserResponse, error) {
	var user User
	collection := c.Database(mongoDatabase).Collection(mongoCollection)
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &pb.GetUserResponse{
		User: mapUserToPbUser(user),
	}, nil
}

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	collection := s.mongoClient.Database(mongoDatabase).Collection(mongoCollection)

	result, err := collection.DeleteOne(ctx, bson.M{"_id": req.GetId()})
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &pb.DeleteUserResponse{
		Success: result.DeletedCount == 1,
	}, nil
}

func mapUserToPbUser(user User) *pb.User {
	return &pb.User{
		Id:   user.Id,
		Name: user.Name,
	}
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
