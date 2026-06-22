import { frontendUrl, oryPublicUrl } from "@/lib/env";
import type { OryFlow, OrySession } from "@/types/ory";

export async function getSession(): Promise<OrySession | null> {
  const response = await fetch(`${oryPublicUrl}/sessions/whoami`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (!response.ok) {
    return null;
  }

  return (await response.json()) as OrySession;
}

export function beginLoginFlow() {
  window.location.href = `${oryPublicUrl}/self-service/login/browser`;
}

export function beginRegistrationFlow() {
  window.location.href = `${oryPublicUrl}/self-service/registration/browser`;
}

export async function beginLogoutFlow() {
  const response = await fetch(
    `${oryPublicUrl}/self-service/logout/browser?return_to=${encodeURIComponent(frontendUrl)}`,
    {
      credentials: "include",
      headers: {
        Accept: "application/json",
      },
    },
  );

  if (!response.ok) {
    throw new Error("Could not start logout flow.");
  }

  const data = (await response.json()) as {
    logout_url: string;
    logout_token: string;
  };

  window.location.href = data.logout_url;
}

export async function fetchLoginFlow(flow: string): Promise<OryFlow> {
  const response = await fetch(
    `${oryPublicUrl}/self-service/login/flows?id=${encodeURIComponent(flow)}`,
    {
      credentials: "include",
      headers: { Accept: "application/json" },
      cache: "no-store",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load login flow from Ory.");
  }

  return (await response.json()) as OryFlow;
}

export async function fetchRegistrationFlow(flow: string): Promise<OryFlow> {
  const response = await fetch(
    `${oryPublicUrl}/self-service/registration/flows?id=${encodeURIComponent(flow)}`,
    {
      credentials: "include",
      headers: { Accept: "application/json" },
      cache: "no-store",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load registration flow from Ory.");
  }

  return (await response.json()) as OryFlow;
}
