#!/bin/bash

# API Testing Script for Online Order Management System
# Make sure the server is running on localhost:8080

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 Testing Online Order Management System API"
echo "=============================================="

# Test 1: Health Check
echo "1. Testing Health Check..."
curl -s -X GET http://localhost:8080/health | jq .
echo -e "\n"

# Test 2: Create a single order
echo "2. Creating a single order..."
ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe",
    "customer_email": "john@example.com",
    "items": [
      {
        "product_name": "Laptop",
        "quantity": 1,
        "unit_price": 999.99
      },
      {
        "product_name": "Mouse",
        "quantity": 2,
        "unit_price": 25.50
      }
    ]
  }')

echo "$ORDER_RESPONSE" | jq .
ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.id')
echo "Created order with ID: $ORDER_ID"
echo -e "\n"

# Test 3: Get the created order
echo "3. Getting order by ID..."
curl -s -X GET "$BASE_URL/orders/$ORDER_ID" | jq .
echo -e "\n"

# Test 4: Update order status
echo "4. Updating order status..."
curl -s -X PUT "$BASE_URL/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "processing"}' | jq .
echo -e "\n"

# Test 5: Create another order for listing test
echo "5. Creating another order..."
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Jane Smith",
    "customer_email": "jane@example.com",
    "items": [
      {
        "product_name": "Keyboard",
        "quantity": 1,
        "unit_price": 75.00
      }
    ]
  }' | jq .
echo -e "\n"

# Test 6: List orders
echo "6. Listing orders..."
curl -s -X GET "$BASE_URL/orders?limit=5" | jq .
echo -e "\n"

# Test 7: Bulk create orders
echo "7. Bulk creating orders..."
curl -s -X POST "$BASE_URL/orders/bulk" \
  -H "Content-Type: application/json" \
  -d '{
    "orders": [
      {
        "customer_name": "Alice Johnson",
        "customer_email": "alice@example.com",
        "items": [
          {
            "product_name": "Monitor",
            "quantity": 1,
            "unit_price": 299.99
          }
        ]
      },
      {
        "customer_name": "Bob Wilson",
        "customer_email": "bob@example.com",
        "items": [
          {
            "product_name": "Headphones",
            "quantity": 1,
            "unit_price": 149.99
          },
          {
            "product_name": "Webcam",
            "quantity": 1,
            "unit_price": 89.99
          }
        ]
      }
    ]
  }' | jq .
echo -e "\n"

# Test 8: List all orders again
echo "8. Listing all orders after bulk creation..."
curl -s -X GET "$BASE_URL/orders?limit=10" | jq .
echo -e "\n"

# Test 9: Test error cases
echo "9. Testing error cases..."

echo "9a. Invalid order ID:"
curl -s -X GET "$BASE_URL/orders/invalid" | jq .
echo -e "\n"

echo "9b. Non-existent order:"
curl -s -X GET "$BASE_URL/orders/99999" | jq .
echo -e "\n"

echo "9c. Invalid status update:"
curl -s -X PUT "$BASE_URL/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "invalid_status"}' | jq .
echo -e "\n"

echo "9d. Invalid order creation (missing required fields):"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "",
    "items": []
  }' | jq .
echo -e "\n"

echo "✅ API Testing Complete!" 