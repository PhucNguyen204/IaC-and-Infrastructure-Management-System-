#!/bin/bash

echo "=========================================="
echo "Running tests for IaaS Management System"
echo "=========================================="

echo ""
echo "1. Testing Authentication Service..."
cd vcs-authentication-service
go test ./... -v -cover
if [ $? -ne 0 ]; then
    echo "Authentication service tests failed"
    exit 1
fi
cd ..

echo ""
echo "2. Testing Infrastructure Provisioning Service..."
cd vcs-infrastructure-provisioning-service
go test ./... -v -cover
if [ $? -ne 0 ]; then
    echo "Provisioning service tests failed"
    exit 1
fi
cd ..

echo ""
echo "3. Testing Infrastructure Monitoring Service..."
cd vcs-infrastructure-monitoring-service
go test ./... -v -cover
if [ $? -ne 0 ]; then
    echo "Monitoring service tests failed"
    exit 1
fi
cd ..

echo ""
echo "=========================================="
echo "All tests passed successfully!"
echo "=========================================="

