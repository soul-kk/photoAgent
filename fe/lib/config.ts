import { Platform } from 'react-native';

/** 后端 API 根地址，可在 .env 设置 EXPO_PUBLIC_API_URL */
export const API_BASE =
  (typeof process !== 'undefined' && process.env?.EXPO_PUBLIC_API_URL) ||
  (Platform.OS === 'android' ? 'http://10.0.2.2:8080' : 'http://localhost:8080');
