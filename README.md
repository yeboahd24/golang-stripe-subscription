# Subscription Application with Stripe Payment Gateway

Developer: Dominic Kofi Yeboah

# Sign up

```bash
curl -X POST http://localhost:8000/register \
-H "Content-Type: application/json" \
-d '{"email": "yeboahd24@gmail.com", "password": "mesika"}'
```

## Response
```json
{
  "message": "User registered successfully"
}
```

# Login

```bash
curl -X POST http://localhost:8000/login \   
-H "Content-Type: application/json" \
-d '{"email": "yeboahd24@gmail.com", "password": "mesika"}'
```

## Response

```json
{
    "token":"TOKEN_HERE"
}
```

# Create Product

NB: Only Admin access

```bash
curl -X POST http://localhost:8000/create-product \
-H "Content-Type: application/json" \                                                            
-H "Authorization: Bearer TOKEN_HERE" \
-d '{
    "name": "Sample Product",
    "description": "This is a sample product description.",
    "monthly_price": 9.99,
    "yearly_price": 99.99
}'
```

## Response

```json
{
    "ID":"34c4b243-c0bf-4c80-ba82-146649ac0eb9",
    "Name":"Sample Product 2",
    "Description":"This is a sample product description.",
    "MonthlyPrice":9.98,
    "YearlyPrice":99.98,
    "StripeMonthlyPriceID":"price_1PsjMiDclBQzaDqrAwjMJasD",
    "StripeYearlyPriceID":"price_1PsjMjDclBQzaDqr0ukWo5EY"
}
```


# Get Subscription

- Plan: monthly or yearly

```bash
curl -X POST http://localhost:8000/subscribe \
-H "Authorization: Bearer TOKEN_HERE" \
-H "Content-Type: application/json" \
-d '{
    "product_id": "34c4b243-c0bf-4c80-ba82-146649ac0eb9",
    "plan": "monthly"
}'
```

## Response

```json

{
    "id":"bce2f357-b78b-4316-a862-5ecd0edbd3b2",
    "user_id":"4e6d0baa-22fb-4f72-8a72-3d136218252c",
    "product_id":"6b0d9de0-24de-4ee2-9b95-08ab472b8961",
    "start_date":"2024-08-28T16:52:20.354701+01:00",
    "end_date":"0000-12-31T23:58:45-00:01",
    "trial_end_date":"2024-09-27T16:52:20.354701+01:00",
    "status":"active",
    "plan":"monthly",
    "stripe_id":"sub_1PskBYDclBQzaDqr96ExgQcg",
    "created_at":"2024-08-28T16:52:20.494378+01:00",
    "updated_at":"2024-08-28T16:52:20.494378+01:00",
    "is_in_trial":false
}
```




# Cancel Subscription

NB: You can only cancel subscription you paid for not free trial.

If you tried to cancel free trial you might get an error.

```json
{
    "error":"Active subscription not found"
}
```


```bash
curl -X POST http://localhost:8000/cancel-subscription \
-H "Authorization: Bearer TOKEN_HERE" \
-H "Content-Type: application/json" \
-d '{
    "subscription_id": "9c2ff0c6-f15d-4226-ae9d-39b6bde3444d"
}'
```

## Response

```json
{
    "message":"Subscription cancelled successfully"
}
```


# Trial Subscription

```bash
curl -X POST http://localhost:8000/trial-subscribe \
-H "Authorization: Bearer TOKEN_HERE" \
-H "Content-Type: application/json" \
-d '{
    "product_id": "6b0d9de0-24de-4ee2-9b95-08ab472b8961"
}'
```

## Response

```json

{
    "id":"bce2f357-b78b-4316-a862-5ecd0edbd3b2",
    "user_id":"4e6d0baa-22fb-4f72-8a72-3d136218252c",
    "product_id":"6b0d9de0-24de-4ee2-9b95-08ab472b8961",
    "start_date":"2024-08-28T16:52:20.354701+01:00",
    "end_date":"0000-12-31T23:58:45-00:01",
    "trial_end_date":"2024-09-27T16:52:20.354701+01:00",
    "status":"active",
    "plan":"trial",
    "stripe_id":"",
    "created_at":"2024-08-28T16:52:20.494378+01:00",
    "updated_at":"2024-08-28T16:52:20.494378+01:00",
    "is_in_trial":true
}
```

# Promote User To Admin

```bash
curl -X POST http://localhost:8000/promote-to-admin \
-H "Authorization: Bearer TOKEN_HERE" \
-H "Content-Type: application/json" \
-d '{
    "user_id": "02defa54-e475-45e0-b932-7d99585d5a57"
}'
```

## Response

```json
{
    "message":"User promoted to admin successfully"
}
```

# Stacks
- Gin-gonic
- Go
- UUID
- JWT
- Postgresql
- Gorm
- Stripe-go