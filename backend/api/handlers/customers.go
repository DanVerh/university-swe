package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"fmt"

	"github.com/DanVerh/university-swe/backend/api/db"
	"github.com/DanVerh/university-swe/backend/api/errorHandling"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo collection name
const customerCollection = "customers"

// CustomersHandler handles requests for customers
type CustomersHandler struct{}

// Customer represents a customer in the database
type Customer struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Name    string             `json:"name" bson:"name"`
	Address string             `json:"address" bson:"address"`
}

// CreateCustomer handles POST requests to add a new customer
func (handler *CustomersHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be POST", nil)
		return
	}

	var customer Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Validate required fields
	if customer.Name == "" || customer.Address == "" {
		errorHandling.ThrowError(w, http.StatusBadRequest, "Name and address are required", nil)
		return
	}

	// Assign a new ID
	customer.ID = primitive.NewObjectID()

	db := db.DbConnect()
	defer db.DbDisconnect()
	collection := db.Client.Database(dbName).Collection(customerCollection)

	_, err := collection.InsertOne(nil, customer)
	if err != nil {
		errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to insert customer into database", err)
		return
	}

	log.Printf("Created customer: %v", customer)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

// List handles GET requests to list all customers or search for customers by name
func (customersHandler *CustomersHandler) List(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("customers")

    // Check if a search query parameter is present
    query := r.URL.Query().Get("name")
    var filter bson.M

    if query != "" {
        // Add a case-insensitive search filter for the name field
        filter = bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}}
    } else {
        // If no query, use an empty filter to list all customers
        filter = bson.M{}
    }

    cursor, err := collection.Find(nil, filter)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to retrieve documents from the database", err)
        return
    }
    defer cursor.Close(nil)

    var customers []Customer
    if err := cursor.All(nil, &customers); err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to decode documents", err)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(customers)
}

// GetByID handles GET requests to retrieve a single customer by ID
func (customersHandler *CustomersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/customers/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("customers")

    var customer Customer
    err = collection.FindOne(nil, bson.M{"_id": objectID}).Decode(&customer)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            errorHandling.ThrowError(w, http.StatusNotFound, "No customer found with the given ID", nil)
        } else {
            errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to retrieve customer", err)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(customer)
}

// UpdateByID handles PUT requests to update a customer by ID
func (customersHandler *CustomersHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be PUT", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/customers/")
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
    collection := db.Client.Database(dbName).Collection("customers")

    var updateKeys []string
    for updateKey := range updateBody {
        if updateKey != "name" && updateKey != "address" {
            errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid update field", nil)
            return
        }
        updateKeys = append(updateKeys, updateKey)
    }

    updateResult, err := collection.UpdateByID(nil, objectID, bson.M{"$set": updateBody})
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to update customer", err)
        return
    }
    if updateResult.MatchedCount == 0 {
        errorHandling.ThrowError(w, http.StatusNotFound, "No customer found with the provided ID", nil)
        return
    }

    response := fmt.Sprintf("Customer with id %v fields updated successfully: %v", id, updateKeys)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(response))
}

// DeleteByID handles DELETE requests to delete a customer by ID
func (customersHandler *CustomersHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be DELETE", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/customers/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
	collection := db.Client.Database(dbName).Collection("customers")

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