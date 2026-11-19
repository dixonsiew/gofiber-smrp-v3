package database

import (
    "context"
    "fmt"
    "os"
    "smrp/config"
    "smrp/utils"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

var dbVar *sqlx.DB
var ctx = context.Background()
var clientVar *mongo.Client
var ctxm = context.TODO()

func SetDb(db *sqlx.DB) {
    dbVar = db
}

func GetDb() *sqlx.DB {
    if dbVar == nil {
        ConnectDB()
    }

    return dbVar
}

func GetCtx() context.Context {
    return ctx
}

func ConnectDB() {
    host := os.Getenv("DB_HOST")
    if host == "" {
        host = "localhost"
    }
    name := config.Config("postgres.db")
    connStr := fmt.Sprintf("postgres://postgres:postgres@%s:5432/%s?sslmode=disable", host, name)
    // connStr := fmt.Sprintf("postgres://postgres:postgres@postgres-service.default.svc.cluster.local:5432/%s?sslmode=disable", name)
    db, err := sqlx.Connect("postgres", connStr)
    if err != nil {
        utils.LogError(err)
    } else {
        db.SetMaxOpenConns(10)
        db.SetMaxIdleConns(5)
        db.SetConnMaxLifetime(time.Hour)
        db.SetConnMaxIdleTime(30 * time.Minute)
        SetDb(db)
    }
}

func CloseDB() {
    dbVar.Close()
}

func ConnectMongo() {
    host := os.Getenv("MONGODB_HOST")
    if host == "" {
        host = "localhost"
    }
    s := fmt.Sprintf("mongodb://%s:27017", host)
    clientOptions := options.Client().ApplyURI(s)
    client, err := mongo.Connect(clientOptions)
    if err != nil {
        utils.LogError(err)
    } else {
        clientVar = client
    }
}

func GetMongoClient() *mongo.Client {
    return clientVar
}

func GetMongoCtx() context.Context {
    return ctxm
}

func CloseMongo() {
    clientVar.Disconnect(ctxm)
}
