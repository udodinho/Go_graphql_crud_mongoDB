package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/udodinho/go-graphql/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

func Connect() MongoInstance {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	dbURI := os.Getenv("DB_URI")

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(dbURI).SetServerAPIOptions(serverAPI)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()
	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		log.Fatal("Context error, mongoDB:", err)
	}

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		panic(err)
	}

	db := client.Database("GO-GRAPHQL-JOB")

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!", result)

	return MongoInstance{
		Client: client,
		Db:     db,
	}

}

func (db *MongoInstance) GetJob(id string) *model.JobListing {
	var jobListing model.JobListing

	collection := db.Db.Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	jobId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: jobId}}

	err := collection.FindOne(ctx, filter).Decode(&jobListing)
	fmt.Println("HERE")
	if err != nil {
		log.Fatal("Unable to fetch user:", err)
	}
	fmt.Println("HERE1")

	return &jobListing
}

func (db *MongoInstance) GetJobs() []*model.JobListing {
	var jobListings []*model.JobListing
	query := bson.D{{}}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	cursor, err := db.Db.Collection("jobs").Find(ctx, query)

	if err != nil {
		log.Fatal("Unable to fetch users", err)
	}

	err = cursor.All(ctx, &jobListings)

	if err != nil {
		panic(err)
	}

	return jobListings
}

func (db *MongoInstance) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {

	collection := db.Db.Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	insertedInfo, err := collection.InsertOne(ctx, bson.M{
		"title":       jobInfo.Title,
		"description": jobInfo.Description,
		"company":     jobInfo.Company,
		"url":         jobInfo.URL,
	})

	if err != nil {
		log.Fatal("Unable to create user", err)
	}

	insertedID := insertedInfo.InsertedID.(primitive.ObjectID).Hex()

	result := model.JobListing{
		ID:          insertedID,
		Title:       jobInfo.Title,
		Description: jobInfo.Description,
		Company:     jobInfo.Company,
		URL:         jobInfo.URL,
	}

	return &result
}

func (db *MongoInstance) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing {

	collection := db.Db.Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	updateInfo := bson.M{}

	if jobInfo.Title != nil {
		updateInfo["title"] = jobInfo.Title
	}

	if jobInfo.Description != nil {
		updateInfo["description"] = jobInfo.Description
	}

	if jobInfo.Company != nil {
		updateInfo["company"] = jobInfo.Company
	}

	if jobInfo.URL != nil {
		updateInfo["url"] = jobInfo.URL
	}

	var jobListing *model.JobListing
	id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.D{{Key: "_id", Value: id}}
	updateJob := bson.M{"$set": updateInfo}

	result := collection.FindOneAndUpdate(ctx, filter, updateJob, options.FindOneAndUpdate().SetReturnDocument(1))

	err := result.Decode(&jobListing)

	if err != nil {
		log.Fatal("Unable to decode result ", err)
	}

	return jobListing
}
func (db *MongoInstance) DeleteJobListing(jobId string) *model.DeleteJobResponse {
	collection := db.Db.Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(jobId)

	filter := bson.D{{Key: "_id", Value: id}}

	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		log.Fatal("Unable to decode result", err)
	}

	return &model.DeleteJobResponse{DeleteJobID: jobId}
}
