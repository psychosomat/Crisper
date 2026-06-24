/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{svelte,html,js}', './index.html'],
  theme: {
    extend: {
      colors: {
        app: {
          bg: '#0d0d0d',
          surface: '#161616',
          elevated: '#1f1f1f',
          border: '#2a2a2a',
          'border-hover': '#404040',
        },
        accent: {
          DEFAULT: '#a0ec06',
          dim: '#8bc905',
        },
        text: {
          primary: '#f5f5f5',
          secondary: '#999999',
          muted: '#666666',
        },
        danger: '#ff453a',
        success: '#30d158',
      },
      fontFamily: {
        ui: ['system-ui', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', '"Cascadia Code"', 'monospace'],
      },
      borderRadius: {
        none: '0',
      },
      borderWidth: {
        DEFAULT: '1px',
      },
      maxWidth: {
        app: '640px',
      },
    },
  },
  plugins: [require('@tailwindcss/forms')],
}
