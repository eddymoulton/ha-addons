#!/bin/bash

echo "Building Home Assistant Config History frontend..."

cd frontend
npm run build

echo "Frontend built successfully! The Go server can now serve the updated UI."