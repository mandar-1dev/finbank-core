// api.js — every network call the frontend makes goes through here so we
// never duplicate fetch() boilerplate across pages.
const API_BASE_URL = window.COREBANK_API_BASE || 'http://localhost:8080/api';

class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
  }
}

async function request(method, path, body, extraHeaders) {
  const headers = { 'Content-Type': 'application/json', ...(extraHeaders || {}) };
  let res;
  try {
    res = await fetch(API_BASE_URL + path, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (networkErr) {
    throw new ApiError('Could not reach the CoreBank API. Is the backend running?', 0);
  }

  let payload = null;
  try {
    payload = await res.json();
  } catch (_) {
    // non-JSON body, fall through
  }

  if (!res.ok) {
    const message = (payload && payload.message) || `Request failed (${res.status})`;
    throw new ApiError(message, res.status);
  }
  return payload ? payload.data : null;
}

const api = {
  get: (path) => request('GET', path),
  post: (path, body, headers) => request('POST', path, body, headers),
  put: (path, body) => request('PUT', path, body),
  delete: (path) => request('DELETE', path),
};
