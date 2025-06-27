## Project Setup
**1.** Clone this repo (https method)
```text
git clone https://github.com/NitishB2M/microservices.git
```
**2.** Move to clone directory
```text
cd microservices
```
**3.** Run below command
```text
go mod tidy
```
**4.** Before running this project you need **.env** file. I don't share .env file, I'll guide you to setups skeleton of it.
```text
//first create .env file in base directory(in microservices dir)
//after that copy below code then replace with your value

EMAIL_ACC = use_your_email_address
EMAIL_PASS = passkey
DB_PASS = your_mysql_database_password
DB_USER = your_mysql_db_user
DB_HOST = localhost
DB_PORT = 3307(change_port acc. to your mysql port)
DB_NAME = ecomm(you may change it)
USER_PORT = 8080
PRODUCT_PORT = 8081
CART_PORT = 8082
ORDER_PORT = 8083
PAYMENT_PORT = 8084
PAYMENT_PUBLISHED_KEY = third_party_payment_integration_pub_key
PAYMENT_SECRET_KEY = third_party_payment_integration_sec_key
```
If you don't want to setups email configuration, check where it is used and then remove it. So, you don't get any errors.
Same for payment integration.

**4.** As this project is based on microservice architecture and rest api.<br/>
So, you need to run different services on multiple port(use don't need to mention any port everything is already setup).
<br/>

4.1. Let's run first microservices **Users**:
```text
//go to users/cmd
cd users/cmd
//run main.go file
go run main.go
```
4.2. Same step-4.1 follows for other microservices. You just need to open new terminal and run below command.
```text
//let say you want to run product microservices
cd product/cmd
//run main.go file
go run main.go
```

---
### Project Structure
```plaintext
├── cart/
├── payment/
├── products/
│   ├── cmd/
│   │   └── main.go
│   ├── dbs/
│   │   └── connection.go
│   ├── internal/
│   │   ├── handlers/
│   │   │   └── product-handler.go
│   │   ├── models/
│   │   │   ├── product.go
│   │   │   └── tags.go
│   │   ├── services/
│   │       ├── filters.go
│   │       └── services.go
│   ├── pkg/
│   ├── uploads/
├── shared/
├── users/
```

### Microservices in golang

- Product Service: Manages the product catalog. It handles product information, such as name, description, price, availability, and categories.
- Cart Service: Manages the user’s shopping cart. It handles adding/removing items, updating quantities, and calculating the total price.
- Order Service: Manages the order lifecycle. This includes order creation, updating order status, tracking, and storing order details.
- Payment Service: Handles payments and interacts with external payment gateways to process transactions.
- Invoice Service: Generates invoices after the payment is confirmed and after an order is placed.
- Messaging Service: Handles communication between different microservices or sends notifications (e.g., emails, SMS) to users about their order or payment status.
- Queue Service: Often used in event-driven architectures to handle asynchronous processing (e.g., for sending confirmation emails or notifying external systems).

# Service Interactions Summary

This section outlines the responsibilities and communication patterns between the services in your microservices architecture for an e-commerce system.

| **Service**        | **Actions/Responsibilities**                                                      | **Communicates with**                              | **Type of Communication**         |
|--------------------|----------------------------------------------------------------------------------|---------------------------------------------------|-----------------------------------|
| **Cart Service**    | - Manages cart (add/remove items, calculate total)                               | Order Service, Product Service, Messaging Service | Synchronous (HTTP)                |
| **Product Service** | - Manages product catalog (name, description, price, availability, categories)  | Cart Service, Order Service                       | Synchronous (HTTP)                |
| **Order Service**   | - Creates and manages orders, updates order status, validates product and pricing | Payment Service, Invoice Service, Product Service, Messaging Service | Synchronous (HTTP) / Asynchronous (Event-based) |
| **Payment Service** | - Processes payments, handles payment gateway communication, updates order status | Order Service, Messaging Service                  | Synchronous (HTTP) / Asynchronous (Event-based) |
| **Invoice Service** | - Generates invoices after successful payment                                     | Messaging Service                                  | Synchronous (HTTP) / Asynchronous (Event-based) |
| **Messaging Service** | - Sends notifications to users (e.g., email, SMS)                              | All services (indirectly)                         | Asynchronous (Event-based)        |
| **Queue Service**   | - Handles message brokering, decouples services                                  | All services (as event consumers/producers)        | Asynchronous (Event-based)        |

---

