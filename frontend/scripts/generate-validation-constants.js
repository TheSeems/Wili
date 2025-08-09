#!/usr/bin/env node

import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import yaml from "js-yaml";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Path to the OpenAPI spec
const openApiPath = path.resolve(__dirname, "../../backend/services/wishlist/openapi.yaml");
const outputPath = path.resolve(__dirname, "../src/lib/api/validation-constants.ts");

function extractValidationConstants(spec) {
  const constants = {};

  // Extract from WishlistItemData schema
  const itemDataSchema = spec.components?.schemas?.WishlistItemData;
  if (itemDataSchema?.properties) {
    if (itemDataSchema.properties.name) {
      constants.ITEM_NAME_MIN_LENGTH = itemDataSchema.properties.name.minLength || 1;
      constants.ITEM_NAME_MAX_LENGTH = itemDataSchema.properties.name.maxLength || 300;
    }
    if (itemDataSchema.properties.description) {
      constants.ITEM_DESCRIPTION_MAX_LENGTH =
        itemDataSchema.properties.description.maxLength || 2000;
    }
  }

  // Extract from CreateWishlistRequest schema
  const createWishlistSchema = spec.components?.schemas?.CreateWishlistRequest;
  if (createWishlistSchema?.properties) {
    if (createWishlistSchema.properties.title) {
      constants.WISHLIST_TITLE_MIN_LENGTH = createWishlistSchema.properties.title.minLength || 1;
      constants.WISHLIST_TITLE_MAX_LENGTH = createWishlistSchema.properties.title.maxLength || 200;
    }
    if (createWishlistSchema.properties.description) {
      constants.WISHLIST_DESCRIPTION_MAX_LENGTH =
        createWishlistSchema.properties.description.maxLength || 2000;
    }
  }

  // Extract from CreateWishlistItemRequest schema
  const createItemSchema = spec.components?.schemas?.CreateWishlistItemRequest;
  if (createItemSchema?.properties) {
    if (createItemSchema.properties.type) {
      constants.ITEM_TYPE_MIN_LENGTH = createItemSchema.properties.type.minLength || 1;
      constants.ITEM_TYPE_MAX_LENGTH = createItemSchema.properties.type.maxLength || 50;
    }
  }

  return constants;
}

function generateTypeScriptConstants(constants) {
  const entries = Object.entries(constants);

  let content = `/**
 * Validation constants extracted from OpenAPI specification
 * This file is auto-generated. Do not edit manually.
 * 
 * Generated from: ${path.relative(process.cwd(), openApiPath)}
 * Generated at: ${new Date().toISOString()}
 */

`;

  // Generate individual constants
  entries.forEach(([key, value]) => {
    const description = key
      .toLowerCase()
      .replace(/_/g, " ")
      .replace(/\b\w/g, (l) => l.toUpperCase());
    content += `/** ${description} */\n`;
    content += `export const ${key} = ${value};\n\n`;
  });

  // Generate validation constants object
  content += `/** All validation constants in a single object */\n`;
  content += `export const VALIDATION_CONSTANTS = {\n`;
  entries.forEach(([key, value]) => {
    content += `  ${key}: ${value},\n`;
  });
  content += `} as const;\n\n`;

  // Generate validation functions
  content += `/** Validation helper functions */\n`;
  content += `export const validation = {\n`;
  content += `  /** Validate item name */\n`;
  content += `  isValidItemName: (name: string): boolean => {\n`;
  content += `    const trimmed = name?.trim() || '';\n`;
  content += `    return trimmed.length >= ITEM_NAME_MIN_LENGTH && trimmed.length <= ITEM_NAME_MAX_LENGTH;\n`;
  content += `  },\n\n`;

  content += `  /** Validate item description */\n`;
  content += `  isValidItemDescription: (description?: string): boolean => {\n`;
  content += `    if (!description) return true; // Optional field\n`;
  content += `    return description.length <= ITEM_DESCRIPTION_MAX_LENGTH;\n`;
  content += `  },\n\n`;

  content += `  /** Validate wishlist title */\n`;
  content += `  isValidWishlistTitle: (title: string): boolean => {\n`;
  content += `    const trimmed = title?.trim() || '';\n`;
  content += `    return trimmed.length >= WISHLIST_TITLE_MIN_LENGTH && trimmed.length <= WISHLIST_TITLE_MAX_LENGTH;\n`;
  content += `  },\n\n`;

  content += `  /** Validate wishlist description */\n`;
  content += `  isValidWishlistDescription: (description?: string): boolean => {\n`;
  content += `    if (!description) return true; // Optional field\n`;
  content += `    return description.length <= WISHLIST_DESCRIPTION_MAX_LENGTH;\n`;
  content += `  },\n\n`;

  content += `  /** Get validation error i18n key for item name */\n`;
  content += `  getItemNameErrorKey: (name: string): string | null => {\n`;
  content += `    const trimmed = name?.trim() || '';\n`;
  content += `    if (trimmed.length === 0) return 'items.nameRequired';\n`;
  content += `    if (trimmed.length > ITEM_NAME_MAX_LENGTH) return 'items.nameTooLong';\n`;
  content += `    return null;\n`;
  content += `  },\n\n`;

  content += `  /** Get validation error i18n key for item description */\n`;
  content += `  getItemDescriptionErrorKey: (description?: string): string | null => {\n`;
  content += `    if (!description) return null;\n`;
  content += `    if (description.length > ITEM_DESCRIPTION_MAX_LENGTH) return 'items.descriptionTooLong';\n`;
  content += `    return null;\n`;
  content += `  },\n\n`;

  content += `  /** Get validation error i18n key for wishlist title */\n`;
  content += `  getWishlistTitleErrorKey: (title: string): string | null => {\n`;
  content += `    const trimmed = title?.trim() || '';\n`;
  content += `    if (trimmed.length === 0) return 'wishlists.titleRequired';\n`;
  content += `    if (trimmed.length > WISHLIST_TITLE_MAX_LENGTH) return 'wishlists.titleTooLong';\n`;
  content += `    return null;\n`;
  content += `  },\n\n`;

  content += `  /** Get validation error i18n key for wishlist description */\n`;
  content += `  getWishlistDescriptionErrorKey: (description?: string): string | null => {\n`;
  content += `    if (!description) return null;\n`;
  content += `    if (description.length > WISHLIST_DESCRIPTION_MAX_LENGTH) return 'wishlists.descriptionTooLong';\n`;
  content += `    return null;\n`;
  content += `  },\n`;
  content += `} as const;\n\n`;

  // Generate TypeScript types
  content += `/** Type definitions for validation */\n`;
  content += `export type ValidationConstants = typeof VALIDATION_CONSTANTS;\n`;
  content += `export type ValidationFunctions = typeof validation;\n`;

  return content;
}

async function main() {
  try {
    console.log("🔍 Reading OpenAPI specification...");
    const yamlContent = fs.readFileSync(openApiPath, "utf8");
    const spec = yaml.load(yamlContent);

    console.log("⚙️  Extracting validation constants...");
    const constants = extractValidationConstants(spec);

    console.log("📝 Generating TypeScript constants...");
    const tsContent = generateTypeScriptConstants(constants);

    // Ensure output directory exists
    const outputDir = path.dirname(outputPath);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    console.log("💾 Writing validation constants file...");
    fs.writeFileSync(outputPath, tsContent, "utf8");

    console.log("✅ Successfully generated validation constants!");
    console.log(`📍 Output: ${path.relative(process.cwd(), outputPath)}`);
    console.log("\n📊 Extracted constants:");
    Object.entries(constants).forEach(([key, value]) => {
      console.log(`   ${key}: ${value}`);
    });
  } catch (error) {
    console.error("❌ Error generating validation constants:", error);
    process.exit(1);
  }
}

main();
