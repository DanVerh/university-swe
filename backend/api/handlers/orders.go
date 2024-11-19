package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/DanVerh/university-swe/backend/api/db"
	"github.com/DanVerh/university-swe/backend/api/errorHandling"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo collection name for orders
const ordersCollection = "orders"

// OrdersHandler handles requests for orders
type OrdersHandler struct{}

// Order represents an order in the database
type Order struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Amount   int32              `json:"amount" bson:"amount"`
	Sum      float64            `json:"sum" bson:"sum"`
	Customer primitive.ObjectID `json:"customer" bson:"customer"`
	Status   string             `json:"status" bson:"status"`
	Product  primitive.ObjectID `json:"product" bson:"product"`
}

// Create handles POST requests to create a new order
func (ordersHandler *OrdersHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be POST", nil)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if order.Amount == 0 || order.Customer == primitive.NilObjectID || order.Product == primitive.NilObjectID  {
		errorHandling.ThrowError(w, http.StatusBadRequest, "Missing required fields: amount, customer, product, or status", nil)
		return
	}

	db := db.DbConnect()
	defer db.DbDisconnect()
	collection := db.Client.Database(dbName).Collection("customers")

	var customerExist Customer
	err := collection.FindOne(nil, bson.M{"_id": order.Customer}).Decode(&customerExist)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			errorHandling.ThrowError(w, http.StatusNotFound, "Customer does not exist", nil)
		} else {
			errorHandling.ThrowError(w, http.StatusInternalServerError, "Error checking customer existence", err)
		}
		return
	}

	productsCollection := db.Client.Database(dbName).Collection("products")
	var productExist Product
	err = productsCollection.FindOne(nil, bson.M{"_id": order.Product}).Decode(&productExist)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			errorHandling.ThrowError(w, http.StatusNotFound, "Product does not exist", nil)
		} else {
			errorHandling.ThrowError(w, http.StatusInternalServerError, "Error checking product existence", err)
		}
		return
	}

	order.Sum = productExist.Price * float64(order.Amount)
	order.Status = "pending"

	// Step 5: Set the new ObjectID for the order and insert into the database
	order.ID = primitive.NewObjectID() // Assign a new ObjectID
	_, err = db.Client.Database(dbName).Collection(ordersCollection).InsertOne(nil, order)
	if err != nil {
		errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to create order", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Order created successfully with ID: %v", order.ID.Hex())))
}

// List handles GET requests to list all orders
func (ordersHandler *OrdersHandler) List(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("orders")

    var filter bson.M
	filter = bson.M{}

    cursor, err := collection.Find(nil, filter)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to retrieve documents from the database", err)
        return
    }
    defer cursor.Close(nil)

    var orders []Order
    if err := cursor.All(nil, &orders); err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to decode documents", err)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}

// GetByID handles GET requests to retrieve a single order by ID
func (ordersHandler *OrdersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/orders/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("orders")

    var order Order
    err = collection.FindOne(nil, bson.M{"_id": objectID}).Decode(&order)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            errorHandling.ThrowError(w, http.StatusNotFound, "No order found with the given ID", nil)
        } else {
            errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to retrieve order", err)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

// UpdateByID handles PUT requests to update an order by ID
func (ordersHandler *OrdersHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be PUT", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/orders/")
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
    collection := db.Client.Database(dbName).Collection("orders")

    var updateKeys []string
    for updateKey := range updateBody {
        if updateKey != "status" {
            errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid update field. Only status allowed", nil)
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
func (ordersHandler *OrdersHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be DELETE", nil)
        return
    }

    id := strings.TrimPrefix(r.URL.Path, "/orders/")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusBadRequest, "Invalid ObjectId format", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
	collection := db.Client.Database(dbName).Collection("orders")

    deleteResult, err := collection.DeleteOne(nil, bson.M{"_id": objectID})
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to delete order", err)
        return
    }
    if deleteResult.DeletedCount == 0 {
        errorHandling.ThrowError(w, http.StatusNotFound, fmt.Sprintf("No order found with the provided ID: %v", id), nil)
        return
    }

    response := fmt.Sprintf("Deleted order with ID: %v", id)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(response))
}

// SumDeliveredOrders handles GET requests to calculate the total sum of delivered orders
func (ordersHandler *OrdersHandler) SumDeliveredOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        errorHandling.ThrowError(w, http.StatusMethodNotAllowed, "Invalid request method. Needs to be GET", nil)
        return
    }

    db := db.DbConnect()
    defer db.DbDisconnect()
    collection := db.Client.Database(dbName).Collection("orders")

    // Aggregation pipeline to filter and sum
    pipeline := mongo.Pipeline{
        bson.D{{"$match", bson.D{{"status", "delivered"}}}},
        bson.D{{"$group", bson.D{
            {"_id", nil},
            {"totalSum", bson.D{{"$sum", "$sum"}}},
        }}},
    }

    cursor, err := collection.Aggregate(nil, pipeline)
    if err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to aggregate orders", err)
        return
    }
    defer cursor.Close(nil)

    // Read the aggregation result
    var result []bson.M
    if err := cursor.All(nil, &result); err != nil {
        errorHandling.ThrowError(w, http.StatusInternalServerError, "Failed to decode aggregation result", err)
        return
    }

    // If no results, the total sum is 0
    var totalSum float64
    if len(result) > 0 {
        totalSum, _ = result[0]["totalSum"].(float64)
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{"totalSum": totalSum})
}