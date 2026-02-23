// src/lib/api.ts
import axios from 'axios';
import { toast } from 'sonner';

const BASE_URL = typeof window !== 'undefined' ? 'http://localhost:8080' : 'http://localhost:8080';

const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

let isRefreshing = false;
let pendingQueue: Array<(token: string | null) => void> = [];

function onRefreshed(token: string | null) {
  pendingQueue.forEach((cb) => cb(token));
  pendingQueue = [];
}

// 请求拦截：自动带 token；FormData 时删除 Content-Type 让浏览器带 boundary
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  if (config.data instanceof FormData) {
    delete config.headers['Content-Type'];
  }
  return config;
});

// 响应拦截：401 自动刷新 token；429 提示频繁操作
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const status = error.response?.status;
    const original = error.config || {};

    if (status === 429) {
      if (typeof window !== 'undefined') {
        toast.error('操作过于频繁，请稍后再试');
      }
      return Promise.reject(error);
    }

    if (status === 401 && !original._retry) {
      original._retry = true;

      if (typeof window === 'undefined') {
        return Promise.reject(error);
      }

      const refreshToken = localStorage.getItem('refresh_token');
      if (!refreshToken) {
        localStorage.removeItem('token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('username');
        window.location.href = '/login';
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingQueue.push((token) => {
            if (!token) return reject(error);
            original.headers = original.headers || {};
            original.headers.Authorization = `Bearer ${token}`;
            resolve(api(original));
          });
        });
      }

      isRefreshing = true;
      try {
        const res = await api.post('/refresh', { refresh_token: refreshToken });
        const newToken = res.data?.access_token;
        const newRefresh = res.data?.refresh_token;

        if (!newToken || !newRefresh) {
          throw new Error('refresh failed');
        }

        localStorage.setItem('token', newToken);
        localStorage.setItem('refresh_token', newRefresh);

        onRefreshed(newToken);
        original.headers = original.headers || {};
        original.headers.Authorization = `Bearer ${newToken}`;
        return api(original);
      } catch (e) {
        onRefreshed(null);
        localStorage.removeItem('token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('username');
        window.location.href = '/login';
        return Promise.reject(e);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export default api;

/** 后端静态资源 base，用于头像、文章图片等 */
export function staticUrl(path: string): string {
  if (!path) return '';
  if (path.startsWith('http')) return path;
  return `${BASE_URL}${path.startsWith('/') ? path : '/' + path}`;
}
