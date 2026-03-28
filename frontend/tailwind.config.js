/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#fdf4f6',
          100: '#fce7eb',
          200: '#f8d0da',
          300: '#f2a9ba',
          400: '#e97693',
          500: '#d9466e',
          600: '#c12e5a',
          700: '#a22047',
          800: '#861e3e',
          900: '#721c39',
        },
      },
    },
  },
  plugins: [],
}
