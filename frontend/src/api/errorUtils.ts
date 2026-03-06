/**
 * Extracts a human-readable error message from an API error response.
 * Looks for error.response.data.error or error.response.data.message,
 * falling back to the provided default string.
 */
export function getErrorMessage(error: unknown, fallback: string): string {
  const err = error as { response?: { data?: { error?: string; message?: string } } };
  return err?.response?.data?.error || err?.response?.data?.message || fallback;
}
