// Nexs Observability MongoDB initialization script

// Switch to nexs_test database
db = db.getSiblingDB('nexs_test');

// Create collections for tracer tests
db.createCollection('test_users');
db.createCollection('test_orders');

// Create collections for logger tests
db.createCollection('test_logs');

// Create indexes for better performance
db.test_users.createIndex({ "email": 1 }, { "unique": true });
db.test_users.createIndex({ "created_at": 1 });

db.test_orders.createIndex({ "order_id": 1 }, { "unique": true });
db.test_orders.createIndex({ "user_id": 1 });
db.test_orders.createIndex({ "status": 1 });
db.test_orders.createIndex({ "created_at": 1 });

db.test_logs.createIndex({ "level": 1 });
db.test_logs.createIndex({ "trace_id": 1 });
db.test_logs.createIndex({ "service_name": 1 });
db.test_logs.createIndex({ "timestamp": 1 });

// Insert sample data for testing
db.test_users.insertMany([
    {
        _id: ObjectId(),
        name: "João Silva",
        email: "joao.silva@example.com",
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        name: "Maria Santos", 
        email: "maria.santos@example.com",
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        name: "Pedro Oliveira",
        email: "pedro.oliveira@example.com", 
        created_at: new Date(),
        updated_at: new Date()
    }
]);

// Get user IDs for orders
var joao = db.test_users.findOne({"email": "joao.silva@example.com"});
var maria = db.test_users.findOne({"email": "maria.santos@example.com"});
var pedro = db.test_users.findOne({"email": "pedro.oliveira@example.com"});

db.test_orders.insertMany([
    {
        _id: ObjectId(),
        user_id: joao._id,
        order_id: "ORD-001",
        amount: 149.99,
        payment_method: "credit_card",
        status: "completed",
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        user_id: joao._id,
        order_id: "ORD-002", 
        amount: 299.50,
        payment_method: "pix",
        status: "pending",
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        user_id: maria._id,
        order_id: "ORD-003",
        amount: 89.90,
        payment_method: "debit_card", 
        status: "completed",
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        _id: ObjectId(),
        user_id: pedro._id,
        order_id: "ORD-004",
        amount: 199.99,
        payment_method: "credit_card",
        status: "processing",
        created_at: new Date(),
        updated_at: new Date()
    }
]);

// Create user for nexs application
db.createUser({
    user: "nexs_app",
    pwd: "nexs123",
    roles: [
        { role: "readWrite", db: "nexs_test" }
    ]
});

print("MongoDB initialization completed successfully!");
