# Personal Finance Manager

A robust and feature-rich personal finance management system built with Go and Plaid API.

## Features

### 1. Advanced Budget Management
- Custom budget categories and sub-categories
- Rolling budgets (monthly, quarterly, yearly)
- Zero-based budgeting system
- Percentage-based budget allocation
- Budget vs. actual spending comparisons
- Flexible budget periods (bi-weekly, monthly, custom dates)

### 2. Enhanced Transaction Management
- Split transaction capability
- Bulk transaction categorization
- Custom tags for transactions
- Receipt attachment and storage
- Transaction notes and memories
- Recurring transaction tracking

### 3. Comprehensive Financial Dashboard
- Real-time account balances
- Income vs. expense trends
- Category-wise spending breakdown
- Bill payment calendar
- Net worth tracker
- Custom widgets for personalization

### 4. Goal Planning Tools
- Multiple savings goals tracking
- Progress visualization
- Target date calculations
- Required monthly savings calculator
- Goal priority setting
- Milestone tracking

### 5. Debt Management Features
- Debt snowball/avalanche calculators
- Interest rate comparisons
- Payoff date estimations
- Payment scheduling
- Total interest saved calculations
- Credit utilization tracking

### 6. Investment Portfolio Tracking
- Multi-account aggregation
- Asset allocation visualization
- Dividend tracking
- Investment performance metrics
- Cost basis tracking
- Realized/unrealized gains tracking

### 7. Reports and Analytics
- Customizable reporting periods
- Export functionality (CSV, PDF)
- Spending trend analysis
- Income source breakdown
- Category comparison over time
- Cash flow statements

### 8. Family Finance Features
- Shared account management
- Individual budgets within family
- Expense splitting
- Allowance tracking for kids
- Family member permissions
- Joint goal setting

### 9. Bill Management
- Due date reminders
- Payment confirmation
- Historical payment tracking
- Recurring bill setup
- Bill categorization
- Payment method tracking

### 10. Security Features
- Multi-factor authentication
- Biometric login
- Session management
- Activity logging
- Regular security notifications
- Data encryption

## Technical Architecture

### Backend Structure
```
/api
  /handlers
    - accounts.go
    - transactions.go
    - budgets.go
    - goals.go
    - reports.go
  /models
    - user.go
    - account.go
    - transaction.go
    - budget.go
  /services
    - plaid_service.go
    - notification_service.go
    - report_service.go
  /middleware
    - auth.go
    - logging.go
    - rate_limit.go
```

### API Endpoints
```
/api/v1
  /auth
    POST /login
    POST /register
    POST /logout
  /accounts
    GET /list
    GET /{id}
    POST /link
  /transactions
    GET /list
    POST /create
    PUT /{id}
    POST /categorize
  /budgets
    GET /current
    POST /create
    PUT /{id}
  /goals
    GET /list
    POST /create
    PUT /{id}
```

## API Documentation

The Personal Finance Manager provides a RESTful API for integrating with the system. The API is documented using OpenAPI/Swagger specification.

### API Base URL
```
http://localhost:8080
```

### Authentication
The API uses JWT (JSON Web Token) for authentication. Include the JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

### Core Endpoints

#### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token

#### Categories
- `GET /api/categories` - Get all categories
- `POST /api/categories` - Create a new category
- `GET /api/categories/{id}` - Get category details
- `PUT /api/categories/{id}` - Update a category
- `DELETE /api/categories/{id}` - Delete a category

#### Accounts
- `GET /api/accounts` - Get user accounts
- `POST /api/accounts` - Create a new account
- `GET /api/accounts/{id}` - Get account details
- `PUT /api/accounts/{id}` - Update an account
- `DELETE /api/accounts/{id}` - Delete an account

#### Transactions
- `GET /api/transactions` - Get user transactions
- `POST /api/transactions` - Create a new transaction
- `GET /api/transactions/{id}` - Get transaction details
- `PUT /api/transactions/{id}` - Update a transaction
- `DELETE /api/transactions/{id}` - Delete a transaction

#### Budgets
- `GET /api/budgets` - Get user budgets
- `POST /api/budgets` - Create a new budget
- `GET /api/budgets/{id}` - Get budget details
- `PUT /api/budgets/{id}` - Update a budget
- `DELETE /api/budgets/{id}` - Delete a budget

#### Analytics
- `GET /api/analytics/spending` - Get spending analytics
- `GET /api/analytics/income` - Get income analytics

#### Recurring Transactions
- `GET /api/recurring` - Get recurring transactions
- `POST /api/recurring` - Create a recurring transaction
- `GET /api/recurring/{id}` - Get recurring transaction details
- `PUT /api/recurring/{id}` - Update a recurring transaction
- `DELETE /api/recurring/{id}` - Delete a recurring transaction

#### Metrics
- `GET /api/metrics` - Get financial metrics

#### Notifications
- `GET /api/notifications` - Get user notifications
- `GET /api/notifications/preferences` - Get notification preferences
- `PUT /api/notifications/preferences` - Update notification preferences
- `PUT /api/notifications/{id}/read` - Mark notification as read

### Query Parameters

#### Transaction Filtering
The `/api/transactions` endpoint supports the following query parameters:
- `start_date` (YYYY-MM-DD) - Filter transactions from this date
- `end_date` (YYYY-MM-DD) - Filter transactions until this date
- `category_id` - Filter by category
- `account_id` - Filter by account
- `type` - Filter by transaction type (income, expense, transfer)

### Detailed API Documentation

For detailed API documentation including request/response schemas, example payloads, and all available endpoints, refer to our Swagger documentation at `/api/swagger.yaml`.

### Rate Limiting

The API implements rate limiting to ensure fair usage:
- 100 requests per minute per IP address
- 1000 requests per hour per user

### Error Handling

The API uses standard HTTP status codes and returns error responses in the following format:
```json
{
  "code": 400,
  "message": "Detailed error message"
}
```

Common error codes:
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (invalid or missing token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

## Technology Stack

### Backend
- Go (1.21+)
- Standard library (net/http) for HTTP server
- sqlc for type-safe SQL
- PostgreSQL for database
- Redis for caching

### Infrastructure
- Docker
- Kubernetes (optional)
- AWS (recommended)

### Third-party Services
- Plaid API for bank connections
- SendGrid for email notifications
- Twilio for SMS alerts
- AWS S3 for document storage
- Stripe for premium features

## Getting Started

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 13 or higher
- Redis
- Plaid API credentials
- AWS account (for production deployment)

### Installation
1. Clone the repository
```bash
git clone https://github.com/yourusername/personal-finance-manager.git
cd personal-finance-manager
```

2. Install dependencies
```bash
go mod download
```

3. Set up environment variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run migrations
```bash
go run cmd/migrate/main.go
```

5. Start the server
```bash
go run cmd/server/main.go
```

## Development

### Running Tests
```bash
go test ./...
```

### Code Style
Follow the official Go style guide and use `gofmt` for formatting:
```bash
gofmt -w .
```

## Deployment

### Docker
```bash
docker build -t personal-finance-manager .
docker run -p 8080:8080 personal-finance-manager
```

### Production Considerations
- Set up proper monitoring and logging
- Configure automated backups
- Implement rate limiting
- Set up CDN for static assets
- Configure proper security measures

## Contributing
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License
This project is licensed under the MIT License - see the LICENSE file for details.
