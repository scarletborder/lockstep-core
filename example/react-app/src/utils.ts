import { LockstepClient } from "lockstep-core-client";

export function buildClient(): LockstepClient {
  return new LockstepClient({
    serverUrl: 'https://your-lockstep-server.com',
    safety: {
      // Define your safety options here
    }
  });
}