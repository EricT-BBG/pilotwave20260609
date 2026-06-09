import axios from 'axios';
import { resolveApiBaseConfig } from './runtime';

const { apiUrl, apiVersion } = resolveApiBaseConfig();
const API_TIMEOUT_MS = 15000;
const API_UNAVAILABLE_KEY = 'Alert.ApiUnavailable';

axios.defaults.timeout = API_TIMEOUT_MS;

const isApiNoResponseError = (error) => {
  if (!error) return false;
  if (error.code === 'ECONNABORTED') return true;
  if (String(error.message || '').toLowerCase().includes('timeout')) return true;
  return !error.response;
};

const notifyApiUnavailable = (error) => {
  if (!isApiNoResponseError(error)) return;
  if (typeof window === 'undefined' || typeof window.dispatchEvent !== 'function') return;

  window.dispatchEvent(new CustomEvent('pilotwave-api-error', {
    detail: {
      message: API_UNAVAILABLE_KEY,
    },
  }));
};

axios.interceptors.response.use(
  (response) => response,
  (error) => {
    notifyApiUnavailable(error);
    return Promise.reject(error);
  }
);

const http = axios.create({
  baseURL: `${apiUrl}${apiVersion}`,
  timeout: API_TIMEOUT_MS,
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json'
  }
});

http.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('accessToken');
  if (token) {
    config.headers.Authentication = token;
  }

  return config;
});

export default http;
