/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Salesforce-inspired color palette
        salesforce: {
          blue: '#0176D3',
          'blue-dark': '#014486',
          'blue-light': '#E5F3FF',
          'blue-hover': '#005FB2',
        },
        neutral: {
          'bg-primary': '#F3F2F2',
          'bg-secondary': '#FFFFFF',
          'text-primary': '#3E3E3C',
          'text-secondary': '#706E6B',
          'border': '#DDDBDA',
          'hover': '#F3F2F2',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        'salesforce': '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        'salesforce-md': '0 2px 4px 0 rgba(0, 0, 0, 0.1)',
      },
    },
  },
  plugins: [],
}

