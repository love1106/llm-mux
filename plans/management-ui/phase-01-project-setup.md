# Phase 1: Project Setup

## Context
Initialize the React project with modern tooling and UI libraries for the llm-mux Management UI.

## Overview
Set up a production-ready React application with TypeScript, Vite, Tailwind CSS, and shadcn/ui components. The project will be structured for embedding into the Go server.

## Requirements
- React 18+ with TypeScript for type safety
- Vite for fast development and optimized builds
- Tailwind CSS for utility-first styling
- shadcn/ui for pre-built, accessible components
- Path configuration for `/ui` base URL

## Implementation Steps

### 1. Initialize Vite Project
```bash
cd /workspace/llm-mux
npm create vite@latest ui -- --template react-ts
cd ui
npm install
```

### 2. Configure TypeScript
Update `tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

### 3. Install Dependencies
```bash
# Core dependencies
npm install react-router-dom@6 @tanstack/react-query zustand axios

# UI dependencies
npm install -D tailwindcss postcss autoprefixer
npm install class-variance-authority clsx tailwind-merge lucide-react

# Development dependencies
npm install -D @types/react @types/react-dom @types/node
```

### 4. Setup Tailwind CSS
```bash
npx tailwindcss init -p
```

Update `tailwind.config.js`:
```javascript
export default {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx,js,jsx}"],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
      },
    },
  },
  plugins: [],
}
```

### 5. Setup shadcn/ui
Create `components.json`:
```json
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "default",
  "rsc": false,
  "tsx": true,
  "tailwind": {
    "config": "tailwind.config.js",
    "css": "src/index.css",
    "baseColor": "slate",
    "cssVariables": true
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils"
  }
}
```

Install essential shadcn/ui components:
```bash
npx shadcn@latest add button card table tabs sheet form
npx shadcn@latest add select input label toast dialog
npx shadcn@latest add dropdown-menu navigation-menu
```

### 6. Configure Vite
Update `vite.config.ts`:
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  base: '/ui/',
  build: {
    outDir: '../dist-ui',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          ui: ['@tanstack/react-query', 'zustand', 'axios'],
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
    proxy: {
      '/v1/management': {
        target: 'http://localhost:8317',
        changeOrigin: true,
      }
    }
  }
})
```

### 7. Project Structure
```
ui/
├── src/
│   ├── components/     # Reusable UI components
│   │   ├── ui/         # shadcn/ui components
│   │   └── layout/     # Layout components
│   ├── pages/          # Route pages
│   ├── hooks/          # Custom React hooks
│   ├── lib/            # Utilities and helpers
│   │   ├── api.ts      # API client
│   │   └── utils.ts    # Utility functions
│   ├── stores/         # Zustand stores
│   ├── types/          # TypeScript definitions
│   ├── App.tsx         # Main app component
│   ├── main.tsx        # Entry point
│   └── index.css       # Global styles
├── public/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## Todo List
- [ ] Initialize Vite project with React TypeScript template
- [ ] Configure TypeScript with path aliases
- [ ] Install and configure Tailwind CSS
- [ ] Setup shadcn/ui with component library
- [ ] Configure Vite for /ui base path and proxy
- [ ] Create initial project structure
- [ ] Setup CSS variables for theming
- [ ] Add development scripts to package.json
- [ ] Create .gitignore for node_modules and build artifacts

## Success Criteria
- [ ] Development server runs on http://localhost:5173
- [ ] Tailwind CSS utilities working
- [ ] shadcn/ui components rendering correctly
- [ ] TypeScript compilation without errors
- [ ] Path aliases (@/) working
- [ ] Proxy to Management API functional
- [ ] Build outputs to ../dist-ui directory