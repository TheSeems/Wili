/**
 * Validation constants extracted from OpenAPI specification
 * This file is auto-generated. Do not edit manually.
 * 
 * Generated from: ../backend/services/wishlist/openapi.yaml
 * Generated at: 2025-10-28T19:23:53.952Z
 */

/** Item Name Min Length */
export const ITEM_NAME_MIN_LENGTH = 1;

/** Item Name Max Length */
export const ITEM_NAME_MAX_LENGTH = 300;

/** Item Description Max Length */
export const ITEM_DESCRIPTION_MAX_LENGTH = 2000;

/** Wishlist Title Min Length */
export const WISHLIST_TITLE_MIN_LENGTH = 1;

/** Wishlist Title Max Length */
export const WISHLIST_TITLE_MAX_LENGTH = 200;

/** Wishlist Description Max Length */
export const WISHLIST_DESCRIPTION_MAX_LENGTH = 2000;

/** Item Type Min Length */
export const ITEM_TYPE_MIN_LENGTH = 1;

/** Item Type Max Length */
export const ITEM_TYPE_MAX_LENGTH = 50;

/** All validation constants in a single object */
export const VALIDATION_CONSTANTS = {
  ITEM_NAME_MIN_LENGTH: 1,
  ITEM_NAME_MAX_LENGTH: 300,
  ITEM_DESCRIPTION_MAX_LENGTH: 2000,
  WISHLIST_TITLE_MIN_LENGTH: 1,
  WISHLIST_TITLE_MAX_LENGTH: 200,
  WISHLIST_DESCRIPTION_MAX_LENGTH: 2000,
  ITEM_TYPE_MIN_LENGTH: 1,
  ITEM_TYPE_MAX_LENGTH: 50,
} as const;

/** Validation helper functions */
export const validation = {
  /** Validate item name */
  isValidItemName: (name: string): boolean => {
    const trimmed = name?.trim() || '';
    return trimmed.length >= ITEM_NAME_MIN_LENGTH && trimmed.length <= ITEM_NAME_MAX_LENGTH;
  },

  /** Validate item description */
  isValidItemDescription: (description?: string): boolean => {
    if (!description) return true; // Optional field
    return description.length <= ITEM_DESCRIPTION_MAX_LENGTH;
  },

  /** Validate wishlist title */
  isValidWishlistTitle: (title: string): boolean => {
    const trimmed = title?.trim() || '';
    return trimmed.length >= WISHLIST_TITLE_MIN_LENGTH && trimmed.length <= WISHLIST_TITLE_MAX_LENGTH;
  },

  /** Validate wishlist description */
  isValidWishlistDescription: (description?: string): boolean => {
    if (!description) return true; // Optional field
    return description.length <= WISHLIST_DESCRIPTION_MAX_LENGTH;
  },

  /** Get validation error i18n key for item name */
  getItemNameErrorKey: (name: string): string | null => {
    const trimmed = name?.trim() || '';
    if (trimmed.length === 0) return 'items.nameRequired';
    if (trimmed.length > ITEM_NAME_MAX_LENGTH) return 'items.nameTooLong';
    return null;
  },

  /** Get validation error i18n key for item description */
  getItemDescriptionErrorKey: (description?: string): string | null => {
    if (!description) return null;
    if (description.length > ITEM_DESCRIPTION_MAX_LENGTH) return 'items.descriptionTooLong';
    return null;
  },

  /** Get validation error i18n key for wishlist title */
  getWishlistTitleErrorKey: (title: string): string | null => {
    const trimmed = title?.trim() || '';
    if (trimmed.length === 0) return 'wishlists.titleRequired';
    if (trimmed.length > WISHLIST_TITLE_MAX_LENGTH) return 'wishlists.titleTooLong';
    return null;
  },

  /** Get validation error i18n key for wishlist description */
  getWishlistDescriptionErrorKey: (description?: string): string | null => {
    if (!description) return null;
    if (description.length > WISHLIST_DESCRIPTION_MAX_LENGTH) return 'wishlists.descriptionTooLong';
    return null;
  },
} as const;

/** Type definitions for validation */
export type ValidationConstants = typeof VALIDATION_CONSTANTS;
export type ValidationFunctions = typeof validation;
