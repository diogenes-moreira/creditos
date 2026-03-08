import { useState, useCallback } from "react";
import {
  RecaptchaVerifier,
  signInWithPhoneNumber,
  type ConfirmationResult,
} from "firebase/auth";
import { auth } from "./config";

export function useFirebasePhoneAuth() {
  const [confirmationResult, setConfirmationResult] =
    useState<ConfirmationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const sendSMS = useCallback(
    async (phone: string, recaptchaContainerId: string) => {
      setLoading(true);
      setError(null);
      try {
        const recaptchaVerifier = new RecaptchaVerifier(
          auth,
          recaptchaContainerId,
          { size: "invisible" }
        );
        const result = await signInWithPhoneNumber(
          auth,
          phone,
          recaptchaVerifier
        );
        setConfirmationResult(result);
        return result;
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "Failed to send SMS";
        setError(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  const verifyCode = useCallback(
    async (code: string): Promise<string> => {
      if (!confirmationResult) {
        throw new Error("No confirmation result — call sendSMS first");
      }
      setLoading(true);
      setError(null);
      try {
        const credential = await confirmationResult.confirm(code);
        const idToken = await credential.user.getIdToken();
        return idToken;
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : "Invalid verification code";
        setError(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [confirmationResult]
  );

  return { sendSMS, verifyCode, loading, error };
}
