#!/bin/sh

# This work is licensed under Apache License, Version 2.0 or later.
# Please read and understand latest version of Licence.

echo "🔧 Setting up your new project from template..."

read -p "Enter your new Go module name (e.g., github.com/yourname/project): " NEW_MODULE

if [ -z "$NEW_MODULE" ]; then
  echo "❌ Module name cannot be empty."
  exit 1
fi

OLD_MODULE="github.com/kazimsarikaya/go_react_mui"

echo "📦 Replacing module: $OLD_MODULE → $NEW_MODULE"

# Replace in go.mod
sed -i "s|$OLD_MODULE|$NEW_MODULE|g" go.mod

# Replace in all .go, .sh, .ts, .tsx, .json files (adjust as needed) but not .git and setup.sh
find . -type f \( -name "*.go" -o -name "*.sh" -o -name "*.ts" -o -name "*.tsx" -o -name "*.json" -o -name Containerfile \) \
   ! -name "setup.sh" \
  -exec sed -i "s|$OLD_MODULE|$NEW_MODULE|g" {} +

# Optional: rename the module path in go.work if present
[ -f go.work ] && sed -i "s|$OLD_MODULE|$NEW_MODULE|g" go.work

echo "🧹 Tidying Go modules..."
go mod tidy

# lint the project
echo "🔧 Setting up golangci-lint..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.2
echo "🔍 Linting the project..."
golangci-lint run --fix

echo "✅ Project is now set up as: $NEW_MODULE"

# Install pre-commit hooks
echo "🔗 Installing git hooks..."
mkdir -p .git/hooks
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
cp .githooks/pre-push .git/hooks/pre-push 
chmod +x .git/hooks/pre-push

# notify user that ensure hooks should be runned.
echo "🔗 Ensure to run the pre-commit/pre-push hooks correctly."
