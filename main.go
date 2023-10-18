package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Context example using in http middleware
type ContextKey string

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

var (
	port = "8080"
	db   *gorm.DB
)

const ContextRequestIdKey ContextKey = "requestId"

func main() {
	if len(os.Getenv("APP_PORT")) != 0 {
		port = os.Getenv("APP_PORT")
	}

	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Product{})

	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		log.Println("Init create Products")
		for i := 0; i < 10; i++ {
			db.Create(&Product{Code: fmt.Sprintf("D10%d", i), Price: 100 + uint(i)})
		}
	}

	// Read
	//var product Product
	//db.First(&product, 1)                 // find product with integer primary key
	//db.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	//db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	//db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	//db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	//db.Delete(&product, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/products", productsHandler)
	// Middleware, can be used simply mux
	handler := logging(mux)

	log.Println("App start on :" + port)
	log.Println("Use APP_PORT environment variable")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		rId := rnd.Intn(1000000000)
		ctx := context.WithValue(r.Context(), ContextRequestIdKey, rId)

		start := time.Now()
		log.Printf("[%10d] %s %s", rId, r.Method, r.RequestURI)
		next.ServeHTTP(w, r.WithContext(ctx)) //r
		log.Printf("[%10d] %s %s %s", rId, r.Method, r.URL.Path, time.Since(start))
	})
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var products []Product
		id := r.URL.Query().Get("id")
		rId := r.Context().Value(ContextRequestIdKey).(int)

		if id != "" {
			db.First(&products, id)
		} else {
			db.Find(&products)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
		log.Printf("[%10d] Return products count %d", rId, len(products))
	}

	if r.Method == http.MethodPost {
		var product Product
		rId := r.Context().Value(ContextRequestIdKey).(int)

		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&product)

		db.Save(&product)
		log.Printf("[%10d] Create product", rId)
	}

	if r.Method == http.MethodPut {
		var product Product
		rId := r.Context().Value(ContextRequestIdKey).(int)

		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&product)

		product.ID = 0
		db.Save(&product)

		w.Header().Set("Location", fmt.Sprintf("%s%d", "/products?id=", product.ID))
		w.WriteHeader(http.StatusCreated)
		log.Printf("[%10d] Update product", rId)
	}

	if r.Method == http.MethodDelete {
		rId := r.Context().Value(ContextRequestIdKey).(int)
		id := r.URL.Query().Get("id")
		db.Delete(&Product{}, id)
		log.Printf("[%10d] Delete product with id=%s", rId, id)
	}
}
