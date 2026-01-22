import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          ui: ['@tanstack/react-query', 'zustand', 'axios'],
          charts: ['recharts'],
        }
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 8318,
    proxy: {
      '/v1/management': {
        target: 'http://localhost:8317',
        changeOrigin: true,
      },
      '/v1/messages': {
        target: 'http://localhost:8317',
        changeOrigin: true,
      }
    }
  },
  preview: {
    port: 8318,
    proxy: {
      '/v1/management': {
        target: 'http://localhost:8317',
        changeOrigin: true,
      },
      '/v1/messages': {
        target: 'http://localhost:8317',
        changeOrigin: true,
      }
    }
  }
})
