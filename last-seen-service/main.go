package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
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

var client *mongo.Client

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongo_uri := getEnv("MONGO_URI", "mongodb://root:example@mongo:27017")
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
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

	log.Printf("gRPC server started on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) UpdateLastSeen(ctx context.Context, req *pb.UpdateLastSeenRequest) (*pb.UpdateLastSeenResponse, error) {
	collection := s.mongoClient.Database("chatdb").Collection("last_seen")

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": req.UserId},
		bson.M{"$set": bson.M{
			"last_seen": time.Unix(req.Timestamp, 0),
		}},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &pb.UpdateLastSeenResponse{
		Status: "success",
	}, nil
}

func (s *server) GetLastSeen(ctx context.Context, req *pb.GetLastSeenRequest) (*pb.GetLastSeenResponse, error) {
	collection := s.mongoClient.Database("chatdb").Collection("last_seen")

	var result UserLastSeen
	err := collection.FindOne(ctx, bson.M{"_id": req.UserId}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &pb.GetLastSeenResponse{
		UserId:   result.UserID,
		LastSeen: timestamppb.New(result.LastSeen),
	}, nil
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
