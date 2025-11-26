#!/bin/bash

set -e

# Get auth credentials
eval $(dd-auth --domain dd.datad0g.com --output)
export DD_SITE="datad0g.com"

# Generate unique table name
TIMESTAMP=$(date +%s)
RANDOM_SUFFIX=$(openssl rand -hex 4)
TABLE_NAME="tf_test_evolution_manual_${TIMESTAMP}_${RANDOM_SUFFIX}"

echo "=== Step 1: Create table with test.csv (3 fields: a, b, c) ==="
echo "Environment: ${DD_SITE}"
echo "Table name: ${TABLE_NAME}"
CREATE_RESPONSE=$(curl -s -X POST "https://dd.${DD_SITE}/api/v2/reference-tables/tables" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "DD-CLIENT-SOURCE: synthetics" \
  -d "{
    \"data\": {
      \"type\": \"reference_table\",
      \"attributes\": {
        \"table_name\": \"${TABLE_NAME}\",
        \"description\": \"Test schema evolution\",
        \"source\": \"S3\",
        \"tags\": [\"test:terraform\"],
        \"file_metadata\": {
          \"sync_enabled\": true,
          \"access_details\": {
            \"aws_detail\": {
              \"aws_account_id\": \"924305315327\",
              \"aws_bucket_name\": \"dd-reference-tables-dev-staging\",
              \"file_path\": \"test.csv\"
            }
          }
        },
        \"schema\": {
          \"primary_keys\": [\"a\"],
          \"fields\": [
            {\"name\": \"a\", \"type\": \"STRING\"},
            {\"name\": \"b\", \"type\": \"STRING\"},
            {\"name\": \"c\", \"type\": \"STRING\"}
          ]
        }
      }
    }
  }")

echo "Create response: $CREATE_RESPONSE"

# Extract table ID from response (might be empty, so we'll list tables to find it)
TABLE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')

if [ -z "$TABLE_ID" ]; then
  echo "No ID in create response, listing tables to find it..."
  LIST_RESPONSE=$(curl -s -X GET "https://dd.${DD_SITE}/api/v2/reference-tables/tables" \
    -H "DD-API-KEY: ${DD_API_KEY}" \
    -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
    -H "Accept: application/json" \
    -H "DD-CLIENT-SOURCE: curl")
  
  TABLE_ID=$(echo "$LIST_RESPONSE" | jq -r ".data[] | select(.attributes.table_name == \"${TABLE_NAME}\") | .id" | head -1)
fi

if [ -z "$TABLE_ID" ]; then
  echo "ERROR: Could not find table ID"
  exit 1
fi

echo "Table ID: $TABLE_ID"
echo "Table name: ${TABLE_NAME}"

echo ""
echo "=== Step 2: Update table with test2.csv (4 fields: a, b, c, d) ==="
PATCH_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PATCH "https://dd.${DD_SITE}/api/v2/reference-tables/tables/${TABLE_ID}" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "DD-CLIENT-SOURCE: synthetics" \
  -d "{
    \"data\": {
      \"type\": \"reference_table\",
      \"attributes\": {
        \"description\": \"Test schema evolution\",
        \"file_metadata\": {
          \"sync_enabled\": true,
          \"access_details\": {
            \"aws_detail\": {
              \"aws_account_id\": \"924305315327\",
              \"aws_bucket_name\": \"dd-reference-tables-dev-staging\",
              \"file_path\": \"test2.csv\"
            }
          }
        },
        \"schema\": {
          \"primary_keys\": [\"a\"],
          \"fields\": [
            {\"name\": \"a\", \"type\": \"STRING\"},
            {\"name\": \"b\", \"type\": \"STRING\"},
            {\"name\": \"c\", \"type\": \"STRING\"},
            {\"name\": \"d\", \"type\": \"STRING\"}
          ]
        },
        \"tags\": [\"test:terraform\"]
      }
    }
  }")

HTTP_STATUS=$(echo "$PATCH_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
PATCH_BODY=$(echo "$PATCH_RESPONSE" | sed '/HTTP_STATUS/d')
echo "Patch HTTP status: $HTTP_STATUS"
echo "Patch response body length: $(echo "$PATCH_BODY" | wc -c)"
if [ -z "$PATCH_BODY" ] || [ "$PATCH_BODY" = "" ]; then
  echo "PATCH response body is empty"
else
  echo "Patch response body:"
  echo "$PATCH_BODY" | jq '.' 2>/dev/null || echo "$PATCH_BODY"
fi

echo ""
echo "=== Step 3: Check if PATCH response contains updated schema ==="
PATCH_SCHEMA_FIELDS=$(echo "$PATCH_BODY" | jq -r '.data.attributes.schema.fields | length // 0' 2>/dev/null)
if [ "$PATCH_SCHEMA_FIELDS" != "0" ]; then
  echo "PATCH response schema field count: $PATCH_SCHEMA_FIELDS"
  echo "PATCH response schema fields:"
  echo "$PATCH_BODY" | jq -r '.data.attributes.schema.fields | map(.name) | join(", ")' 2>/dev/null || echo "Could not parse"
else
  echo "PATCH response has no schema data (empty body or different format)"
fi

echo ""
echo "=== Step 4: Immediately read back (no delay) ==="
PATCH_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
IMMEDIATE_RESPONSE=$(curl -s -X GET "https://dd.${DD_SITE}/api/v2/reference-tables/tables/${TABLE_ID}" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
  -H "Accept: application/json" \
  -H "DD-CLIENT-SOURCE: synthetics")
IMMEDIATE_FIELDS=$(echo "$IMMEDIATE_RESPONSE" | jq -r '.data.attributes.schema.fields | map(.name) | join(", ")')
IMMEDIATE_UPDATED=$(echo "$IMMEDIATE_RESPONSE" | jq -r '.data.attributes.updated_at')
IMMEDIATE_FIELD_COUNT=$(echo "$IMMEDIATE_RESPONSE" | jq -r '.data.attributes.schema.fields | length')
echo "Immediate read - Field count: $IMMEDIATE_FIELD_COUNT"
echo "Immediate read - Fields: $IMMEDIATE_FIELDS"
echo "Immediate read - Updated at: $IMMEDIATE_UPDATED"
echo "PATCH completed at: $PATCH_TIME"

echo ""
echo "=== Waiting 3 seconds before read ==="
sleep 3

echo ""
echo "=== Step 5: Read back the table (after 3 second delay) ==="
AFTER_DELAY_RESPONSE=$(curl -s -X GET "https://dd.${DD_SITE}/api/v2/reference-tables/tables/${TABLE_ID}" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
  -H "Accept: application/json" \
  -H "DD-CLIENT-SOURCE: synthetics")
AFTER_DELAY_FIELDS=$(echo "$AFTER_DELAY_RESPONSE" | jq -r '.data.attributes.schema.fields | map(.name) | join(", ")')
AFTER_DELAY_UPDATED=$(echo "$AFTER_DELAY_RESPONSE" | jq -r '.data.attributes.updated_at')
AFTER_DELAY_FIELD_COUNT=$(echo "$AFTER_DELAY_RESPONSE" | jq -r '.data.attributes.schema.fields | length')
echo "After delay - Field count: $AFTER_DELAY_FIELD_COUNT"
echo "After delay - Fields: $AFTER_DELAY_FIELDS"
echo "After delay - Updated at: $AFTER_DELAY_UPDATED"
echo ""
echo "Full response:"
echo "$AFTER_DELAY_RESPONSE" | jq '.'

echo ""
echo "=== Cleanup: Delete table ==="
curl -s -X DELETE "https://dd.${DD_SITE}/api/v2/reference-tables/tables/${TABLE_ID}" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
  -H "Accept: application/json" \
  -H "DD-CLIENT-SOURCE: synthetics" > /dev/null
echo "Table deleted"

