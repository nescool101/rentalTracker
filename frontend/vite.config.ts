import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  // Configure base path for GitHub Pages - deploy to root since that's where it's working
  base: '/',
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
      manifest: {
        name: 'Rental Management System',
        short_name: 'RentManager',
        description: 'A comprehensive rental property management application',
        theme_color: '#ffffff',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }
        ]
      }
    })
  ],
  build: {
    // Improve chunk size and reduce the number of chunks
    rollupOptions: {
      output: {
        manualChunks: {
          // Group React and related dependencies
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          // Group UI framework dependencies
          'mantine-vendor': [
            '@mantine/core',
            '@mantine/hooks',
            '@mantine/form',
            '@mantine/notifications',
            '@mantine/dates'
          ],
          // Group utility libraries
          'utils-vendor': ['axios', '@tanstack/react-query']
        },
        // Ensure chunks have meaningful names
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]'
      }
    },
    // Enable source maps for production builds
    sourcemap: false,
    // Decrease chunks
    cssCodeSplit: false
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      }
    }
  }
})
