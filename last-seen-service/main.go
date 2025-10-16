package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "last-seen-service/pb"
)

type UserLastSeen struct {
	UserID   string    `bson:"_id"`
	LastSeen time.Time `bson:"last_seen"`
}

type server struct {
	pb.UnimplementedLastSeenServiceServer
	mongoClient *mongo.Client
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongo_uri := getEnv("MONGO_URI", "mongodb://root:example@mongo:27017")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Trying to connect to MongoDB.")
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	log.Println("Successfully connected to MongoDB!")

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	lastSeenServer := &server{
		mongoClient: client,
	}

	pb.RegisterLastSeenServiceServer(s, lastSeenServer)

	enable_reflection := getEnv("REFLECTION", "false")
	log.Println("Reflection enabled: ", enable_reflection)
	if enable_reflection == "true" {
		reflection.Register(s)
	}

	log.Println("gRPC server started on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) UpdateLastSeen(ctx context.Context, req *pb.UpdateLastSeenRequest) (*pb.UpdateLastSeenResponse, error) {
	collection := s.mongoClient.Database("chatdb").Collection("last_seen")

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": req.GetUserId()},
		bson.M{"$set": bson.M{
			"last_seen": time.Unix(req.GetTimestamp(), 0),
		}},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		log.Printf("failed to update last seen for user %s: %v", req.GetUserId(), err)
		return nil, status.Errorf(codes.Internal, "failed to update last seen")
	}

	return &pb.UpdateLastSeenResponse{}, nil
}

func (s *server) GetLastSeen(ctx context.Context, req *pb.GetLastSeenRequest) (*pb.GetLastSeenResponse, error) {
	collection := s.mongoClient.Database("chatdb").Collection("last_seen")

	var result UserLastSeen
	err := collection.FindOne(ctx, bson.M{"_id": req.GetUserId()}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("user not found: %s", req.GetUserId())
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Printf("database error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve last seen")
	}

	return &pb.GetLastSeenResponse{
		LastSeen: timestamppb.New(result.LastSeen),
	}, nil
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
