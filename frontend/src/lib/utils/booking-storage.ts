const STORAGE_KEY = "wili_booking_tokens";

interface BookingToken {
  wishlistId: string;
  itemId: string;
  cancellationToken: string;
  bookedAt: string;
}

function getTokens(): BookingToken[] {
  if (typeof window === "undefined") return [];
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? JSON.parse(stored) : [];
  } catch {
    return [];
  }
}

function saveTokens(tokens: BookingToken[]): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens));
  } catch (error) {
    console.error("Failed to save booking tokens:", error);
  }
}

export function saveBookingToken(
  wishlistId: string,
  itemId: string,
  cancellationToken: string
): void {
  const tokens = getTokens();
  const filtered = tokens.filter((t) => !(t.wishlistId === wishlistId && t.itemId === itemId));
  filtered.push({
    wishlistId,
    itemId,
    cancellationToken,
    bookedAt: new Date().toISOString(),
  });
  saveTokens(filtered);
}

export function getBookingToken(wishlistId: string, itemId: string): string | null {
  const tokens = getTokens();
  const token = tokens.find((t) => t.wishlistId === wishlistId && t.itemId === itemId);
  return token ? token.cancellationToken : null;
}

export function removeBookingToken(wishlistId: string, itemId: string): void {
  const tokens = getTokens();
  const filtered = tokens.filter((t) => !(t.wishlistId === wishlistId && t.itemId === itemId));
  saveTokens(filtered);
}

export function hasBookingToken(wishlistId: string, itemId: string): boolean {
  return getBookingToken(wishlistId, itemId) !== null;
}

