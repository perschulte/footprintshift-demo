#!/bin/bash

echo "Testing GreenWeb Impact API"
echo "========================="

# Start the server in background
export PORT=8092
./greenweb > /dev/null 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "1. Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8092/health)
echo "Health: $HEALTH"

echo ""
echo "2. Testing video streaming calculation..."
VIDEO_RESULT=$(curl -s -X POST http://localhost:8092/api/v1/impact/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "type": "video_streaming",
    "duration": 3600,
    "video_quality": "1080p",
    "device_type": "laptop",
    "connection_type": "wifi",
    "region": "EU",
    "optimization_level": 30,
    "include_rebound_effects": true
  }')

if echo "$VIDEO_RESULT" | grep -q "baseline_emissions"; then
    echo "✅ Video streaming calculation works!"
    echo "   Baseline emissions: $(echo "$VIDEO_RESULT" | grep -o '"baseline_emissions":[0-9.]*' | cut -d: -f2)"
    echo "   Savings: $(echo "$VIDEO_RESULT" | grep -o '"savings":[0-9.]*' | cut -d: -f2)g CO2"
else
    echo "❌ Video streaming calculation failed"
    echo "Response: $VIDEO_RESULT"
fi

echo ""
echo "3. Testing image optimization calculation..."
IMAGE_RESULT=$(curl -s -X POST http://localhost:8092/api/v1/impact/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "type": "image_loading",
    "image_count": 25,
    "data_size": 12.5,
    "device_type": "smartphone",
    "connection_type": "mobile_4g",
    "region": "US",
    "optimization_level": 70
  }')

if echo "$IMAGE_RESULT" | grep -q "baseline_emissions"; then
    echo "✅ Image optimization calculation works!"
    echo "   Savings percentage: $(echo "$IMAGE_RESULT" | grep -o '"savings_percentage":[0-9.]*' | cut -d: -f2)%"
else
    echo "❌ Image optimization calculation failed"
    echo "Response: $IMAGE_RESULT"
fi

echo ""
echo "4. Testing AI inference calculation..."
AI_RESULT=$(curl -s -X POST http://localhost:8092/api/v1/impact/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "type": "ai_inference",
    "duration": 10.0,
    "device_type": "laptop",
    "connection_type": "wifi",
    "region": "US",
    "optimization_level": 50
  }')

if echo "$AI_RESULT" | grep -q "baseline_emissions"; then
    echo "✅ AI inference calculation works!"
    echo "   Data center percentage: $(echo "$AI_RESULT" | grep -o '"datacenter_percentage":[0-9.]*' | cut -d: -f2)%"
else
    echo "❌ AI inference calculation failed"
    echo "Response: $AI_RESULT"
fi

echo ""
echo "5. Testing validation endpoint..."
VALIDATION_RESULT=$(curl -s -X POST http://localhost:8092/api/v1/impact/validate \
  -H "Content-Type: application/json" \
  -d '{
    "claimed_savings": 150.5,
    "optimization_type": "image_loading",
    "parameters": {
      "image_count": 30,
      "data_size": 15.0
    }
  }')

if echo "$VALIDATION_RESULT" | grep -q "validated_savings"; then
    echo "✅ Validation endpoint works!"
    echo "   Rating: $(echo "$VALIDATION_RESULT" | grep -o '"rating":"[^"]*"' | cut -d: -f2 | tr -d '"')"
    echo "   Is valid: $(echo "$VALIDATION_RESULT" | grep -o '"is_valid":[^,]*' | cut -d: -f2)"
else
    echo "❌ Validation endpoint failed"
    echo "Response: $VALIDATION_RESULT"
fi

echo ""
echo "6. Testing dashboard endpoint..."
DASHBOARD_RESULT=$(curl -s http://localhost:8092/api/v1/impact/dashboard)

if echo "$DASHBOARD_RESULT" | grep -q "methodology"; then
    echo "✅ Dashboard endpoint works!"
    echo "   Contains real-time metrics and methodology info"
else
    echo "❌ Dashboard endpoint failed"
    echo "Response: $DASHBOARD_RESULT"
fi

echo ""
echo "7. Testing examples endpoint..."
EXAMPLES_RESULT=$(curl -s http://localhost:8092/api/v1/impact/examples)

if echo "$EXAMPLES_RESULT" | grep -q "video_streaming_calculation"; then
    echo "✅ Examples endpoint works!"
    echo "   Contains calculation examples and methodology"
else
    echo "❌ Examples endpoint failed"
fi

# Stop the server
kill $SERVER_PID 2>/dev/null

echo ""
echo "========================="
echo "Impact API testing complete!"