// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    host: true,  // 监听所有接口（0.0.0.0 + ::），兼容 127.0.0.1 和 localhost
    port: 3000,
    proxy: {
      '/admin/api': {
        target: 'http://localhost:9080',
        changeOrigin: true,
      },
    },
  },
})
