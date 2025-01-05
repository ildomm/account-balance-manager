#!/bin/bash

# Target URLs
BASE_URL="http://localhost:8080"
POST_URL_TEMPLATE="$BASE_URL/user/{USER_ID}/transaction"  # Template for different users
GET_URL_TEMPLATE="$BASE_URL/user/{USER_ID}/balance"  # Template for getting balance of specific user

# User IDs for balance check
USER_IDS=(1 2 3 4 5 6 7 8 9 10 11 12)

# Number of transactions for each user
NUM_TRANSACTIONS=100

# Amount per transaction
TRANSACTION_AMOUNT=1.11

# Function for POSTing transactions
post_transaction() {
  user_id=$1
  amount=$2
  state="win"
  txn_id=$(uuidgen)

  # Substitute user_id into the POST_URL
  POST_URL=${POST_URL_TEMPLATE//\{USER_ID\}/$user_id}

  curl -X POST "$POST_URL" \
    -H "Content-Type: application/json" \
    -H "Source-Type: game" \
    -d "{\"state\": \"$state\", \"amount\": \"$amount\", \"transactionId\": \"$txn_id\"}" \
    -s -o /dev/null -w "%{http_code}"
}

# Function to GET user balance
get_balance() {
  user_id=$1
  # Substitute user_id into the GET_URL
  GET_URL=${GET_URL_TEMPLATE//\{USER_ID\}/$user_id}

  curl -X GET "$GET_URL" -s | jq '.balance'
}

# Load test function to apply transactions concurrently
load_test() {
  for ((i = 0; i < NUM_TRANSACTIONS; i++)); do
    # Fire off POST requests for each user concurrently
    for user_id in "${USER_IDS[@]}"; do
      post_transaction "$user_id" "$TRANSACTION_AMOUNT" &
    done
    # Wait for all background jobs to complete before continuing with the next iteration
    wait
  done
}

# Run the load test
echo "Starting load test for $NUM_TRANSACTIONS transactions per user..."
load_test
echo "Load test completed."

# Fetch final balances for users and compare with expected (1110)
echo "Final balances:"
for user_id in "${USER_IDS[@]}"; do
  final_balance=$(get_balance "$user_id")
  echo "User $user_id: $final_balance"
done

# Check if the final balance matches the expected value (1110)
echo "Checking if final balances match the expected value..."
for user_id in "${USER_IDS[@]}"; do
  final_balance=$(get_balance "$user_id")
done
