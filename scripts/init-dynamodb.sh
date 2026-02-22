#!/bin/sh
set -e

echo "Waiting for LocalStack to be ready..."
until curl -s http://localstack:4566/_localstack/health | grep -q '"dynamodb": "available"'; do
  sleep 1
done

echo "Creating DynamoDB table: urls"
aws --endpoint-url=http://localstack:4566 \
    --region ap-northeast-1 \
    dynamodb create-table \
    --table-name urls \
    --attribute-definitions AttributeName=code,AttributeType=S \
    --key-schema AttributeName=code,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    2>&1 | grep -v "Table already exists" || true

echo "DynamoDB table ready."
