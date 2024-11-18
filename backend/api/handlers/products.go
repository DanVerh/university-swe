package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/DanVerh/university-swe/backend/api/db"
	"github.com/DanVerh/university-swe/backend/api/errorHandling"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo collection name
const dbName = "sales"

// ProductsHandler handles requests for products
type ProductsHandler struct{}

// Product represents a product in the database
type Product struct {
    ID    primitive.ObjectID `json:"id" bson:"_id"`
    Name  string             `json:"name" bson:"name"`
    Price float64            `json:"price" bson:"price"`
    Amount *int32            `json:"amount" bson:"amount"`
}

func (productHandler *ProductsHandler) Create(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be POST", nil)
        return
    }

    var product Product
    d := json.NewDecoder(r.Body)
    d.UseNumber()

    if err := d.Decode(&product); err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid JSON", nil)
        return
    }

    if product.Name == "" || product.Price <= 0 {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Name is required and price must be positive", nil)
        return
    }

    product.ID = primitive.NewObjectID()
    amount := int32(0)
	product.Amount = &amount

    db := db.DbConnect()
    if db == nil {
        log.Fatal("Failed to connect to the database")
    }
    defer db.DbDisconnect()

    collection := db.Client.Database(dbName).Collection("products")

    // Optional: Log the product before insertion
    log.Printf("Product to insert: %+v", product)

    _, err := collection.InsertOne(nil, product)
    if err != nil {
        log.Printf("Failed to insert product into the database: %v", err)
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to insert the product into the database", err)
        return
    }

    log.Printf("Created product: %v, %v\n", product.Name, product.Price)

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

// List handles GET requests to list all products
func (productHandler *ProductsHandler) List(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("products")

    cursor, err := collection.Find(nil, bson.M{})
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to retrieve documents from the database", err)
        return
    }
    defer cursor.Close(nil)

    var products []Product
    if err := cursor.All(nil, &products); err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to decode documents", err)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

// GetByID handles GET requests to retrieve a single product by ID
func (productHandler *ProductsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/products/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("products")

    var product Product
    err = collection.FindOne(nil, bson.M{"_id": objectID}).Decode(&product)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            errorHandling.ThrowError(w, http.StatusNotFound, "No product found with the given ID", nil)
        } else {
            errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to retrieve product", err)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

// UpdateByID handles PUT requests to update a product by ID
func (productHandler *ProductsHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be PUT", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/products/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    var updateBody bson.M
    if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid request body", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("products")

    var updateKeys []string
    for updateKey, updateValue := range updateBody {
        if updateKey != "name" && updateKey != "price" && updateKey != "amount" {
            errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid update field", nil)
            return
        }
        if updateKey == "amount" {
            floatValue, ok := updateValue.(float64)
            if !ok {
                errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid type for 'amount'. Expected a number.", nil)
                return
            }
            updateBody[updateKey] = int32(floatValue)
        }
        updateKeys = append(updateKeys, updateKey)
    }

    updateResult, err := collection.UpdateByID(nil, objectID, bson.M{"$set": updateBody})
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to update product", err)
        return
    }
    if updateResult.MatchedCount == 0 {
        errorHandling.ThrowError(w, http.StatusNotFound, "No product found with the provided ID", nil)
        return
    }

    response := fmt.Sprintf("Product with id %v fields updated successfully: %v", id, updateKeys)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(response))
}

// DeleteByID handles DELETE requests to delete a product by ID
func (productHandler *ProductsHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be DELETE", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/products/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
	collection := db.Client.Database(dbName).Collection("products")

    deleteResult, err := collection.DeleteOne(nil, bson.M{"_id": objectID})
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to delete product", err)
        return
    }
    if deleteResult.DeletedCount == 0 {
        errorHandling.ThrowError(w, http.StatusNotFound, fmt.Sprintf("No product found with the provided ID: %v", id), nil)
        return
    }

    response := fmt.Sprintf("Deleted product with ID: %v", id)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(response))
}

// SearchProducts handles GET requests to search for products by name
func (productHandler *ProductsHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    query := r.URL.Query().Get("name")
    if query == "" {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Search query is required", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("products")

    filter := bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}}
    cursor, err := collection.Find(nil, filter)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to search for products", err)
        return
    }
    defer cursor.Close(nil)

    var products []Product
    if err := cursor.All(nil, &products); err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to decode products", err)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}
