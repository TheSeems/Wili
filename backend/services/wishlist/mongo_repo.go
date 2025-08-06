package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	openapi_types "github.com/oapi-codegen/runtime/types"
	wishlistgen "github.com/theseems/wili/backend/services/wishlist/gen"
)

type MongoRepo struct {
	client    *mongo.Client
	db        *mongo.Database
	wishlists *mongo.Collection
}

type mongoWishlist struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"`
	UUID        string              `bson:"uuid"` // Store the actual UUID
	UserID      string              `bson:"userId"`
	Title       string              `bson:"title"`
	Description *string             `bson:"description"`
	Items       []mongoWishlistItem `bson:"items"`
	CreatedAt   time.Time           `bson:"createdAt"`
	UpdatedAt   time.Time           `bson:"updatedAt"`
}

type mongoWishlistItem struct {
	ID        string                 `bson:"id"`
	Type      string                 `bson:"type"`
	Data      map[string]interface{} `bson:"data"`
	CreatedAt time.Time              `bson:"createdAt"`
	UpdatedAt time.Time              `bson:"updatedAt"`
}

func NewMongoRepo(uri, dbName string) (*MongoRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(dbName)
	wishlists := db.Collection("wishlists")

	// Create indexes
	_, err = wishlists.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create userId index: %w", err)
	}

	// Create UUID index for fast lookups
	_, err = wishlists.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create uuid index: %w", err)
	}

	return &MongoRepo{
		client:    client,
		db:        db,
		wishlists: wishlists,
	}, nil
}

// Helper functions for UUID lookup
func (r *MongoRepo) findByUUID(ctx context.Context, wishlistUUID openapi_types.UUID) (*mongoWishlist, error) {
	var mw mongoWishlist
	err := r.wishlists.FindOne(ctx, bson.M{"uuid": wishlistUUID.String()}).Decode(&mw)
	if err != nil {
		return nil, err
	}
	return &mw, nil
}

func (r *MongoRepo) CreateWishlist(ctx context.Context, userID openapi_types.UUID, req wishlistgen.CreateWishlistRequest) (*wishlistgen.Wishlist, error) {
	now := time.Now()
	wishlistUUID := uuid.New() // Generate a proper UUID

	doc := mongoWishlist{
		UUID:        wishlistUUID.String(),
		UserID:      userID.String(),
		Title:       req.Title,
		Description: req.Description,
		Items:       []mongoWishlistItem{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := r.wishlists.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert wishlist: %w", err)
	}

	return &wishlistgen.Wishlist{
		Id:          openapi_types.UUID(wishlistUUID),
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		Items:       []wishlistgen.WishlistItem{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (r *MongoRepo) GetWishlistsByUser(ctx context.Context, userID openapi_types.UUID) ([]wishlistgen.Wishlist, error) {
	filter := bson.M{"userId": userID.String()}
	cursor, err := r.wishlists.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find wishlists: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoWishlists []mongoWishlist
	if err = cursor.All(ctx, &mongoWishlists); err != nil {
		return nil, fmt.Errorf("failed to decode wishlists: %w", err)
	}

	wishlists := make([]wishlistgen.Wishlist, len(mongoWishlists))
	for i, mw := range mongoWishlists {
		wishlists[i] = r.convertToAPIWishlist(mw)
	}

	return wishlists, nil
}

func (r *MongoRepo) GetWishlistByID(ctx context.Context, wishlistID openapi_types.UUID) (*wishlistgen.Wishlist, error) {
	mw, err := r.findByUUID(ctx, wishlistID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find wishlist: %w", err)
	}

	wishlist := r.convertToAPIWishlist(*mw)
	return &wishlist, nil
}

func (r *MongoRepo) DeleteWishlist(ctx context.Context, wishlistID openapi_types.UUID, userID openapi_types.UUID) error {
	filter := bson.M{
		"uuid":   wishlistID.String(),
		"userId": userID.String(),
	}

	result, err := r.wishlists.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete wishlist: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("wishlist not found or not owned by user")
	}

	return nil
}

func (r *MongoRepo) AddItemToWishlist(ctx context.Context, wishlistID openapi_types.UUID, userID openapi_types.UUID, req wishlistgen.CreateWishlistItemRequest) (*wishlistgen.WishlistItem, error) {
	now := time.Now()
	itemID := uuid.New()

	item := mongoWishlistItem{
		ID:        itemID.String(),
		Type:      req.Type,
		Data:      req.Data,
		CreatedAt: now,
		UpdatedAt: now,
	}

	filter := bson.M{
		"uuid":   wishlistID.String(),
		"userId": userID.String(),
	}

	update := bson.M{
		"$push": bson.M{"items": item},
		"$set":  bson.M{"updatedAt": now},
	}

	result, err := r.wishlists.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to wishlist: %w", err)
	}

	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("wishlist not found or not owned by user")
	}

	return &wishlistgen.WishlistItem{
		Id:        itemID,
		Type:      req.Type,
		Data:      req.Data,
		CreatedAt: &now,
		UpdatedAt: &now,
	}, nil
}

func (r *MongoRepo) UpdateWishlistItem(ctx context.Context, wishlistID, itemID openapi_types.UUID, userID openapi_types.UUID, req wishlistgen.UpdateWishlistItemRequest) (*wishlistgen.WishlistItem, error) {
	now := time.Now()

	filter := bson.M{
		"uuid":     wishlistID.String(),
		"userId":   userID.String(),
		"items.id": itemID.String(),
	}

	update := bson.M{
		"$set": bson.M{
			"items.$.updatedAt": now,
			"updatedAt":         now,
		},
	}

	if req.Type != nil {
		update["$set"].(bson.M)["items.$.type"] = *req.Type
	}
	if req.Data != nil {
		update["$set"].(bson.M)["items.$.data"] = *req.Data
	}

	result, err := r.wishlists.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update wishlist item: %w", err)
	}

	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("wishlist or item not found, or not owned by user")
	}

	// Return updated item (simplified)
	return &wishlistgen.WishlistItem{
		Id:        itemID,
		Type:      *req.Type,
		Data:      *req.Data,
		UpdatedAt: &now,
	}, nil
}

func (r *MongoRepo) DeleteWishlistItem(ctx context.Context, wishlistID, itemID openapi_types.UUID, userID openapi_types.UUID) error {
	now := time.Now()

	filter := bson.M{
		"uuid":   wishlistID.String(),
		"userId": userID.String(),
	}

	update := bson.M{
		"$pull": bson.M{"items": bson.M{"id": itemID.String()}},
		"$set":  bson.M{"updatedAt": now},
	}

	result, err := r.wishlists.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete wishlist item: %w", err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("wishlist or item not found, or not owned by user")
	}

	return nil
}

func (r *MongoRepo) UpdateWishlist(ctx context.Context, wishlistID openapi_types.UUID, userID openapi_types.UUID, title string, description *string) error {
	filter := bson.M{
		"uuid":   wishlistID.String(),
		"userId": userID.String(),
	}

	update := bson.M{
		"$set": bson.M{
			"title":     title,
			"updatedAt": time.Now(),
		},
	}

	if description != nil {
		update["$set"].(bson.M)["description"] = *description
	}

	result, err := r.wishlists.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update wishlist: %w", err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("wishlist not found or not owned by user")
	}

	return nil
}

func (r *MongoRepo) convertToAPIWishlist(mw mongoWishlist) wishlistgen.Wishlist {
	// Use the stored UUID
	wishlistID := openapi_types.UUID(uuid.MustParse(mw.UUID))
	userID := uuid.MustParse(mw.UserID)

	items := make([]wishlistgen.WishlistItem, len(mw.Items))
	for i, item := range mw.Items {
		itemID := uuid.MustParse(item.ID)
		items[i] = wishlistgen.WishlistItem{
			Id:        itemID,
			Type:      item.Type,
			Data:      item.Data,
			CreatedAt: &item.CreatedAt,
			UpdatedAt: &item.UpdatedAt,
		}
	}

	return wishlistgen.Wishlist{
		Id:          wishlistID,
		UserId:      userID,
		Title:       mw.Title,
		Description: mw.Description,
		Items:       items,
		CreatedAt:   mw.CreatedAt,
		UpdatedAt:   mw.UpdatedAt,
	}
}

func (r *MongoRepo) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
