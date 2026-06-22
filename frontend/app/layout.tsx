import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Sukoon Image Studio',
  description: 'Ory + Next.js + MinIO + Go + OpenAI image storytelling frontend.',
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
