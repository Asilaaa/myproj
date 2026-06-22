'use client';

import { useEffect, useMemo, useState } from 'react';

import {
  deleteImage,
  fetchBucketObjects,
  fetchImages,
  fetchMe,
  importImage,
  uploadImage,
  type BucketObject,
  type ImageRecord,
} from '@/lib/backend';
import { beginLoginFlow, beginLogoutFlow, getSession } from '@/lib/ory';
import type { OrySession } from '@/types/ory';

const pastelMessages = [
  'A little spark of vision magic is ready for you.',
  'Your private gallery is humming with possibility.',
  'Upload, explore, describe — let’s make your bucket joyful.',
];

export function Dashboard() {
  const [session, setSession] = useState<OrySession | null>(null);
  const [backendUserID, setBackendUserID] = useState<string | null>(null);
  const [images, setImages] = useState<ImageRecord[]>([]);
  const [objects, setObjects] = useState<BucketObject[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const loadAll = async () => {
    try {
      setLoading(true);
      setError(null);

      const currentSession = await getSession();
      setSession(currentSession);

      if (!currentSession) {
        setImages([]);
        setObjects([]);
        setBackendUserID(null);
        return;
      }

      const [me, nextImages, nextObjects] = await Promise.all([
        fetchMe(),
        fetchImages(),
        fetchBucketObjects(),
      ]);

      setBackendUserID(me.user_id);
      setImages(nextImages);
      setObjects(nextObjects);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong while loading your dashboard.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadAll();
  }, []);

  const filteredImages = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return images;
    return images.filter((image) =>
      [image.original_filename, image.object_name, image.description]
        .join(' ')
        .toLowerCase()
        .includes(needle),
    );
  }, [images, query]);

  const importableObjects = useMemo(() => {
    const described = new Set(images.map((image) => image.object_name));
    return objects.filter((object) => !described.has(object.name));
  }, [images, objects]);

  const heroMessage = pastelMessages[images.length % pastelMessages.length];

  async function handleUpload() {
    if (!selectedFile) {
      setError('Choose an image first.');
      return;
    }

    try {
      setBusy('upload');
      setError(null);
      setSuccess(null);
      await uploadImage(selectedFile);
      setSelectedFile(null);
      setSuccess('Image uploaded and described successfully ✨');
      await loadAll();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed.');
    } finally {
      setBusy(null);
    }
  }

  async function handleImport(objectName: string) {
    try {
      setBusy(`import:${objectName}`);
      setError(null);
      setSuccess(null);
      await importImage(objectName);
      setSuccess(`Imported ${objectName} and generated a fresh description.`);
      await loadAll();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed.');
    } finally {
      setBusy(null);
    }
  }

  async function handleDelete(id: string) {
    try {
      setBusy(`delete:${id}`);
      setError(null);
      setSuccess(null);
      await deleteImage(id);
      setSuccess('Image removed from your gallery and bucket.');
      await loadAll();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Delete failed.');
    } finally {
      setBusy(null);
    }
  }

  return (
    <main className="app-shell">
      <div className="orb orb--pink" />
      <div className="orb orb--violet" />
      <div className="orb orb--gold" />

      <section className="hero-panel glass-card">
        <div>
          <span className="pill">Joyful image intelligence</span>
          <h1>Sukoon Image Studio</h1>
          <p>{heroMessage}</p>
          {session ? (
            <div className="hero-meta">
              <span>Signed in as <strong>{session.identity.traits?.email ?? session.identity.id}</strong></span>
              <span>Backend user id: <strong>{backendUserID ?? 'loading…'}</strong></span>
            </div>
          ) : (
            <div className="hero-meta">
              <span>No active session yet — let’s fix that in one click.</span>
            </div>
          )}
        </div>

        <div className="hero-actions">
          {session ? (
            <button className="candy-button candy-button--secondary" onClick={() => beginLogoutFlow()}>
              Sign out
            </button>
          ) : (
            <button className="candy-button" onClick={() => beginLoginFlow()}>
              Sign in with Ory
            </button>
          )}
          <button className="ghost-button" onClick={() => void loadAll()} disabled={loading || !!busy}>
            Refresh dashboard
          </button>
        </div>
      </section>

      {error && <div className="toast toast--error">{error}</div>}
      {success && <div className="toast toast--success">{success}</div>}

      {!session ? (
        <section className="welcome-grid">
          <article className="glass-card showcase-card">
            <h2>Authenticate first</h2>
            <p>
              Your frontend is ready, your backend is wired, and your images are waiting. Sign in through Ory to unlock your personal MinIO bucket and AI-powered descriptions.
            </p>
            <button className="candy-button" onClick={() => beginLoginFlow()}>
              Launch the login flow
            </button>
          </article>
        </section>
      ) : loading ? (
        <section className="stats-grid">
          <div className="glass-card shimmer-card">Loading your gallery…</div>
          <div className="glass-card shimmer-card">Gathering bucket objects…</div>
          <div className="glass-card shimmer-card">Mixing colors and metadata…</div>
        </section>
      ) : (
        <>
          <section className="stats-grid">
            <article className="stat-card glass-card">
              <span className="stat-card__label">Described images</span>
              <strong>{images.length}</strong>
            </article>
            <article className="stat-card glass-card">
              <span className="stat-card__label">Objects in bucket</span>
              <strong>{objects.length}</strong>
            </article>
            <article className="stat-card glass-card">
              <span className="stat-card__label">Ready to import</span>
              <strong>{importableObjects.length}</strong>
            </article>
          </section>

          <section className="workspace-grid">
            <article className="glass-card panel-card">
              <div className="section-heading">
                <span className="pill">Upload</span>
                <h2>Send a fresh image</h2>
                <p>Drop in a new image and let the backend store it in MinIO, then ask OpenAI for a thoughtful description.</p>
              </div>

              <label className="upload-dropzone">
                <input
                  type="file"
                  accept="image/*"
                  onChange={(event) => setSelectedFile(event.target.files?.[0] ?? null)}
                />
                <div>
                  <strong className="truncate-2" title={selectedFile?.name ?? 'Choose an image file'}>
                    {selectedFile ? selectedFile.name : 'Choose an image file'}
                  </strong>
                  <span>{selectedFile ? `${Math.round(selectedFile.size / 1024)} KB selected` : 'PNG, JPG, WEBP — bright memories welcome.'}</span>
                </div>
              </label>

              <button className="candy-button" onClick={() => void handleUpload()} disabled={busy === 'upload'}>
                {busy === 'upload' ? 'Uploading magic…' : 'Upload and describe'}
              </button>
            </article>

            <article className="glass-card panel-card">
              <div className="section-heading">
                <span className="pill">Bucket explorer</span>
                <h2>Bring existing objects into the gallery</h2>
                <p>These images already live in your personal MinIO bucket. Import one to create a DB record and AI description.</p>
              </div>

              <div className="bucket-list">
                {importableObjects.length === 0 ? (
                  <p className="muted-copy">Every image object is already represented in your described gallery. Nice work.</p>
                ) : (
                  importableObjects.map((object) => (
                    <div className="bucket-item" key={object.name}>
                      <div>
                        <strong className="truncate-2" title={object.name}>{object.name}</strong>
                        <span>{object.content_type || 'image'} · {Math.round(object.size / 1024)} KB</span>
                      </div>
                      <button
                        className="ghost-button"
                        onClick={() => void handleImport(object.name)}
                        disabled={busy === `import:${object.name}`}
                      >
                        {busy === `import:${object.name}` ? 'Importing…' : 'Import'}
                      </button>
                    </div>
                  ))
                )}
              </div>
            </article>
          </section>

          <section className="gallery-section glass-card">
            <div className="gallery-toolbar">
              <div className="section-heading">
                <span className="pill">Gallery</span>
                <h2>Your described images</h2>
                <p>Search by file name, object path, or description and keep the prettiest insights close.</p>
              </div>

              <input
                className="search-input"
                placeholder="Search your descriptions…"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
            </div>

            <div className="gallery-grid">
              {filteredImages.length === 0 ? (
                <div className="empty-card">
                  <h3>No described images yet</h3>
                  <p>Upload something colorful or import an existing bucket object to get started.</p>
                </div>
              ) : (
                filteredImages.map((image) => (
                  <article className="image-card" key={image.id}>
                    <div className="image-card__top">
                      <span className="badge">{image.content_type}</span>
                      <button
                        className="icon-button"
                        onClick={() => void handleDelete(image.id)}
                        disabled={busy === `delete:${image.id}`}
                        aria-label={`Delete ${image.original_filename}`}
                      >
                        {busy === `delete:${image.id}` ? '…' : '✕'}
                      </button>
                    </div>
                    <h3 className="truncate-2" title={image.original_filename}>{image.original_filename}</h3>
                    <p className="image-card__meta truncate-2" title={image.object_name}>{image.object_name}</p>
                    <p className="image-card__description" title={image.description}>{image.description}</p>
                    <div className="image-card__footer">
                      <span>{Math.round(image.size_bytes / 1024)} KB</span>
                      <span>{new Date(image.created_at).toLocaleString()}</span>
                    </div>
                  </article>
                ))
              )}
            </div>
          </section>
        </>
      )}
    </main>
  );
}
