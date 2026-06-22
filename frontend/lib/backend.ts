import { backendProxyBase } from '@/lib/env';

export type ImageRecord = {
  id: string;
  user_id: string;
  object_name: string;
  original_filename: string;
  content_type: string;
  size_bytes: number;
  description: string;
  created_at: string;
};

export type BucketObject = {
  name: string;
  size: number;
  content_type: string;
  last_modified: string;
};

async function readJson<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let message = 'Request failed';
    try {
      const error = (await response.json()) as { error?: string };
      if (error.error) message = error.error;
    } catch {
      // ignore parse failure
    }
    throw new Error(message);
  }

  return (await response.json()) as T;
}

export async function fetchMe() {
  const response = await fetch(`${backendProxyBase}/api/me`, {
    credentials: 'include',
    cache: 'no-store',
  });
  return readJson<{ user_id: string }>(response);
}

export async function fetchImages() {
  const response = await fetch(`${backendProxyBase}/api/images`, {
    credentials: 'include',
    cache: 'no-store',
  });
  return readJson<ImageRecord[]>(response);
}

export async function fetchBucketObjects() {
  const response = await fetch(`${backendProxyBase}/api/objects`, {
    credentials: 'include',
    cache: 'no-store',
  });
  return readJson<BucketObject[]>(response);
}

export async function uploadImage(file: File) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch(`${backendProxyBase}/api/images/upload`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  return readJson<ImageRecord>(response);
}

export async function importImage(objectName: string) {
  const response = await fetch(`${backendProxyBase}/api/images/import`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ object_name: objectName }),
  });

  return readJson<ImageRecord>(response);
}

export async function deleteImage(id: string) {
  const response = await fetch(`${backendProxyBase}/api/images/${id}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    let message = 'Delete failed';
    try {
      const error = (await response.json()) as { error?: string };
      if (error.error) message = error.error;
    } catch {
      // ignore parse failure
    }
    throw new Error(message);
  }
}
