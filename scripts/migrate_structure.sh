#!/bin/bash
# Migration script for restructuring gedcom-go
# Moves files and updates package declarations

set -e

ROOT_DIR="/apps/gedcom-go"
cd "$ROOT_DIR"

echo "Starting migration..."

# Function to update package declaration in a file
update_package() {
    local file=$1
    local new_package=$2
    sed -i "s/^package gedcom$/package $new_package/" "$file"
    sed -i "s/^package parser$/package $new_package/" "$file"
    sed -i "s/^package validator$/package $new_package/" "$file"
    sed -i "s/^package exporter$/package $new_package/" "$file"
    sed -i "s/^package query$/package $new_package/" "$file"
    sed -i "s/^package diff$/package $new_package/" "$file"
    sed -i "s/^package duplicate$/package $new_package/" "$file"
}

# Function to move file and update package
move_and_update() {
    local src=$1
    local dst=$2
    local pkg=$3
    if [ -f "$src" ]; then
        mkdir -p "$(dirname "$dst")"
        cp "$src" "$dst"
        update_package "$dst" "$pkg"
        echo "Moved: $src -> $dst"
    fi
}

echo "Step 1: Moving types files..."
# Move types files (excluding query, diff, duplicate subdirectories)
for file in pkg/gedcom/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "types/$filename" "types"
    fi
done

echo "Step 2: Moving parser files..."
for file in internal/parser/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "parser/$filename" "parser"
    fi
done

echo "Step 3: Moving validator files..."
for file in internal/validator/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "validator/$filename" "validator"
    fi
done

echo "Step 4: Moving exporter files..."
for file in internal/exporter/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "exporter/$filename" "exporter"
    fi
done

echo "Step 5: Moving query files..."
for file in pkg/gedcom/query/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "query/$filename" "query"
    fi
done

echo "Step 6: Moving diff files..."
for file in pkg/gedcom/diff/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "diff/$filename" "diff"
    fi
done

echo "Step 7: Moving duplicate files..."
for file in pkg/gedcom/duplicate/*.go; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        move_and_update "$file" "duplicate/$filename" "duplicate"
    fi
done

echo "Migration complete! Files moved. Next: update imports."





