#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Generating user-service kitex code..."
echo "kitex -module example.com/micro-auth-demo/user-service -service UserService ${ROOT_DIR}/idl/user.thrift"

echo "Generating auth-service kitex code..."
echo "kitex -module example.com/micro-auth-demo/auth-service -service AuthService ${ROOT_DIR}/idl/auth.thrift"

echo "Generation commands prepared. Run them after installing kitex."
