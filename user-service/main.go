package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	authpb "user-service/auth-pb"
	pb "user-service/pb"
)

type User struct {
	Id       uuid.UUID `bson:"_id"`
	Name     string    `bson:"name"`
	Username string    `bson:"username"`
}

type server struct {
	pb.UnimplementedUserServiceServer
	authClient  authpb.AuthServiceClient
	mongoClient *mongo.Client
}

const mongoDatabase = "chatdb"
const mongoCollection = "user"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	var client *mongo.Client

	authServiceUrl := getEnv("AUTH_SERVICE_URL", "auth-service:50051")
	conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	defer conn.Close()

	mongoUri := getEnv("MONGO_URI", "mongodb://root:example@mongo:27017")
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Trying to connect to MongoDB.")
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	log.Println("Successfully connected to MongoDB!")

	// username field should be unique
	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	_, err = collection.Indexes().CreateOne(
		context.TODO(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	userServer := &server{
		authClient:  authpb.NewAuthServiceClient(conn),
		mongoClient: client,
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userServer)

	enableReflection := getEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enableReflection)
	if enableReflection == "true" {
		reflection.Register(s)
	}

	log.Println("gRPC server started on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	session, err := s.mongoClient.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start mongo session: %w", err)
	}
	defer session.EndSession(ctx)

	collection := s.mongoClient.Database(mongoDatabase).Collection(mongoCollection)

	userInsert := User{
		Id:       uuid.New(),
		Name:     req.GetName(),
		Username: req.GetUsername(),
	}

	var insertedId string

	txnFunc := func(sessCtx mongo.SessionContext) (any, error) {
		result, err := collection.InsertOne(sessCtx, userInsert)
		if err != nil {
			return nil, err
		}
		insertedId = result.InsertedID.(string)

		_, err = s.authClient.RegisterCredentials(sessCtx, &authpb.RegisterCredentialsRequest{
			UserId:   insertedId,
			Password: req.GetPassword(),
		})
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, txnFunc)
	if err != nil {
		return nil, fmt.Errorf("transaction failed to commit or was aborted: %w", err)
	}

	return &pb.CreateUserResponse{Id: insertedId}, nil
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
		Id:       user.Id.String(),
		Username: user.Username,
		Name:     user.Name,
	}
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
