# Frontend

Next.js frontend for the Ory + MinIO + Go + OpenAI image studio.

## Environment

Copy `.env.example` to `.env.local` and adjust values:

```bash
cp .env.example .env.local
```

Required variables:
- `NEXT_PUBLIC_FRONTEND_URL` – URL where the Next.js app runs
- `NEXT_PUBLIC_ORY_PUBLIC_URL` – Ory public/browser endpoint
- `BACKEND_URL` – Go backend base URL reachable from the Next.js server

## Important Ory note

For the custom login and registration pages to work, Ory should be configured so that its UI URLs point to this frontend, for example:
- login UI URL -> `http://localhost:3000/login`
- registration UI URL -> `http://localhost:3000/registration`

The pages then load the Ory flow and submit the form directly to Ory.

## Development

```bash
npm install
npm run dev
```

Backend should be running separately, and the Next.js app proxies backend requests through `/api/backend/*`.
